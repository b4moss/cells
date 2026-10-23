/*
 * Copyright (c) 2018. Abstrium SAS <team (at) pydio.com>
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

package grpc

import "time"

// effectiveConsentRequestedAt picks the clock used by CreateConsent's
// ConsentRequestMaxAge check. A zero RequestedAt (seen in production as
// fosite request_unauthorized at handler CreateConsent) would make
// zero.Add(maxAge).Before(now) always true; fall back to AuthenticatedAt
// (set to now by AcceptLogin) or now. Issue #14 / patch 006.
func effectiveConsentRequestedAt(requestedAt, authenticatedAt, now time.Time) (effective time.Time, usedFallback bool) {
	if !requestedAt.IsZero() {
		return requestedAt, false
	}
	if !authenticatedAt.IsZero() {
		return authenticatedAt, true
	}
	return now, true
}

// consentLoginExpired reports whether the login request is outside maxAge.
func consentLoginExpired(requestedAt, authenticatedAt time.Time, maxAge time.Duration, now time.Time) (expired bool, effective time.Time, usedFallback bool) {
	effective, usedFallback = effectiveConsentRequestedAt(requestedAt, authenticatedAt, now)
	return effective.Add(maxAge).Before(now), effective, usedFallback
}
