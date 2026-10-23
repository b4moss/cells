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

import (
	"testing"
	"time"
)

func TestEffectiveConsentRequestedAt(t *testing.T) {
	now := time.Date(2026, 10, 23, 12, 6, 45, 0, time.UTC)
	authAt := now.Add(-time.Second)
	reqAt := now.Add(-2 * time.Minute)

	t.Run("uses RequestedAt when set", func(t *testing.T) {
		got, fb := effectiveConsentRequestedAt(reqAt, authAt, now)
		if fb || !got.Equal(reqAt) {
			t.Fatalf("got %v fallback=%v", got, fb)
		}
	})
	t.Run("falls back to AuthenticatedAt when RequestedAt zero", func(t *testing.T) {
		got, fb := effectiveConsentRequestedAt(time.Time{}, authAt, now)
		if !fb || !got.Equal(authAt) {
			t.Fatalf("got %v fallback=%v", got, fb)
		}
	})
	t.Run("falls back to now when both zero", func(t *testing.T) {
		got, fb := effectiveConsentRequestedAt(time.Time{}, time.Time{}, now)
		if !fb || !got.Equal(now) {
			t.Fatalf("got %v fallback=%v", got, fb)
		}
	})
}

func TestConsentLoginExpired(t *testing.T) {
	now := time.Date(2026, 10, 23, 12, 6, 45, 0, time.UTC)
	maxAge := 30 * time.Minute

	t.Run("zero RequestedAt without fallback would always expire; with AuthenticatedAt does not", func(t *testing.T) {
		// Legacy bug: zero.Add(30m).Before(now) == true
		var zero time.Time
		if !zero.Add(maxAge).Before(now) {
			t.Fatal("precondition: zero RequestedAt must look expired with naive check")
		}
		expired, _, fb := consentLoginExpired(time.Time{}, now.Add(-time.Second), maxAge, now)
		if expired || !fb {
			t.Fatalf("expired=%v fallback=%v", expired, fb)
		}
	})
	t.Run("genuinely stale RequestedAt expires", func(t *testing.T) {
		stale := now.Add(-time.Hour)
		expired, eff, fb := consentLoginExpired(stale, now, maxAge, now)
		if !expired || fb || !eff.Equal(stale) {
			t.Fatalf("expired=%v eff=%v fallback=%v", expired, eff, fb)
		}
	})
	t.Run("fresh RequestedAt ok", func(t *testing.T) {
		fresh := now.Add(-time.Minute)
		expired, _, fb := consentLoginExpired(fresh, now, maxAge, now)
		if expired || fb {
			t.Fatalf("expired=%v fallback=%v", expired, fb)
		}
	})
}
