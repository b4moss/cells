/*
 * Copyright (c) 2021. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package modifiers

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"

	"github.com/pydio/cells/v5/common/auth"
	"github.com/pydio/cells/v5/common/auth/claim"
	"github.com/pydio/cells/v5/common/auth/hydra"
	"github.com/pydio/cells/v5/common/proto/rest"
	"github.com/pydio/cells/v5/common/service/frontend"
	"github.com/pydio/cells/v5/common/telemetry/log"
	"github.com/pydio/cells/v5/idm/oauth"
)

// LoginExternalAuth allows users having a valid Cells session to create an authorization code directly.
// If the FrontSession access_token is missing or expired, refresh_token is used (Issue #7 / 36h re-SSO).
func LoginExternalAuth(middleware frontend.AuthMiddleware) frontend.AuthMiddleware {
	return func(req *restful.Request, rsp *restful.Response, in *frontend.FrontSessionWithRuntimeCtx, out *rest.FrontSessionResponse, session *sessions.Session) error {
		if a, ok := in.AuthInfo["type"]; !ok || a != "external" { // Ignore this middleware
			return middleware(req, rsp, in, out, session)
		}

		challenge, ok := in.AuthInfo["challenge"]
		if !ok {
			return errors.New("Challenge is required")
		}

		ctx := req.Request.Context()
		jwtVerifier := auth.DefaultJWTVerifier()

		var claims claim.Claims
		accessToken, hasAccess := session.Values["access_token"].(string)
		if hasAccess && accessToken != "" {
			var err error
			_, claims, err = jwtVerifier.Verify(ctx, accessToken)
			if err != nil {
				hasAccess = false
			}
		} else {
			hasAccess = false
		}

		if !hasAccess {
			refreshToken, hasRefresh := session.Values["refresh_token"].(string)
			if !hasRefresh || refreshToken == "" {
				return errors.New("Access token is required")
			}
			token, rerr := hydra.Refresh(ctx, refreshToken)
			if rerr != nil {
				log.Logger(ctx).Warn("LoginExternalAuth refresh failed", zap.Error(rerr))
				return rerr
			}
			session.Values["access_token"] = token.AccessToken
			session.Values["id_token"] = token.IDToken
			session.Values["refresh_token"] = token.RefreshToken
			session.Values["expires_at"] = strconv.Itoa(int(time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Unix()))
			accessToken = token.AccessToken
			var err error
			_, claims, err = jwtVerifier.Verify(ctx, accessToken)
			if err != nil {
				return err
			}
		}

		login, code, err := jwtVerifier.LoginChallengeCode(ctx, claims, auth.SetChallenge(challenge))
		if err != nil {
			// Issue #14 / patch 006 (b): do not mutate FrontSession tokens on
			// CreateConsent/LoginChallengeCode failure — caller keeps the Cells
			// session so a failed Gitea SSO does not force password re-entry
			// for file.b4m.jp itself (logout is a separate AuthInfo type).
			log.Logger(ctx).Warn("LoginExternalAuth LoginChallengeCode failed", zap.Error(err))
			return err
		}

		requestURL, err := url.Parse(login.GetRequestURL())
		if err != nil {
			return err
		}

		requestURLValues := requestURL.Query()

		redirectURL, err := oauth.GetRedirectURIFromRequestValues(requestURLValues)
		if err != nil {
			return err
		}

		out.RedirectTo = redirectURL + "?code=" + code + "&state=" + requestURLValues.Get("state")

		// Sliding 36h remember on each successful re-SSO.
		extendFrontSessionRemember(session)
		if cerr := setHydraRememberCookie(req, rsp, login.GetSessionID()); cerr != nil {
			log.Logger(ctx).Warn("setHydraRememberCookie failed", zap.Error(cerr))
		}

		return middleware(req, rsp, in, out, session)
	}
}
