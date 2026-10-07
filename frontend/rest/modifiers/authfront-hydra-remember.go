/*
 * Copyright (c) 2021. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Spike helper: set Hydra authentication-session cookie after login/consent
 * so re-SSO can skip the password UI for RememberFor seconds.
 */

package modifiers

import (
	"net/http"

	restful "github.com/emicklei/go-restful/v3"
	"github.com/gorilla/sessions"
	"github.com/ory/hydra/v2/consent"

	"github.com/pydio/cells/v5/common/runtime/manager"
	"github.com/pydio/cells/v5/idm/oauth"
)

// oauthLoginRememberFor is 36h in seconds (PO requirement).
const oauthLoginRememberFor = 36 * 60 * 60

// setHydraRememberCookie writes ory_hydra_session with sid=sessionID.
// CreateAuthCode uses a nil ResponseWriter so Hydra never sets this cookie;
// FrontSession must set it after a successful challenge login.
func setHydraRememberCookie(req *restful.Request, rsp *restful.Response, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	ctx := req.Request.Context()
	reg, err := manager.Resolve[oauth.Registry](ctx)
	if err != nil {
		return err
	}
	store, err := reg.CookieStore(ctx)
	if err != nil {
		return err
	}
	cookie, _ := store.Get(req.Request, reg.Config().SessionCookieName(ctx))
	cookie.Values[consent.CookieAuthenticationSIDName] = sessionID
	cookie.Options.MaxAge = oauthLoginRememberFor
	cookie.Options.HttpOnly = true
	cookie.Options.Path = reg.Config().SessionCookiePath(ctx)
	cookie.Options.SameSite = reg.Config().CookieSameSiteMode(ctx)
	cookie.Options.Secure = reg.Config().CookieSecure(ctx)
	return cookie.Save(req.Request, rsp.ResponseWriter)
}

func extendFrontSessionRemember(session *sessions.Session) {
	if session.Options == nil {
		session.Options = &sessions.Options{}
	}
	session.Options.MaxAge = oauthLoginRememberFor
	session.Options.HttpOnly = true
	if session.Options.SameSite == 0 {
		session.Options.SameSite = http.SameSiteLaxMode
	}
}
