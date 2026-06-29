package tokens

import (
	"strings"
	"testing"
	"time"

	authdomain "learnflow_backend/internal/auth/domain"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/smartystreets/goconvey/convey"
)

func newTokens(secret, prevSecret, issuer, audience string) *Tokens {
	return NewTokens(secret, prevSecret, issuer, audience)
}

func TestMakeHash(t *testing.T) {
	Convey("MakeHash", t, func() {
		Convey("returns 64-char hex string", func() {
			So(len(MakeHash("test")), ShouldEqual, 64)
		})

		Convey("is deterministic", func() {
			So(MakeHash("hello"), ShouldEqual, MakeHash("hello"))
		})

		Convey("different inputs produce different hashes", func() {
			So(MakeHash("a"), ShouldNotEqual, MakeHash("b"))
		})
	})
}

func TestGenerateSecureToken(t *testing.T) {
	Convey("GenerateSecureToken", t, func() {
		raw, hash, err := GenerateSecureToken()

		Convey("returns no error", func() {
			So(err, ShouldBeNil)
		})

		Convey("raw is 43 chars (32 bytes base64url no-padding)", func() {
			So(len(raw), ShouldEqual, 43)
		})

		Convey("hash is 64-char hex SHA-256", func() {
			So(len(hash), ShouldEqual, 64)
		})

		Convey("hash matches MakeHash(raw)", func() {
			So(hash, ShouldEqual, MakeHash(raw))
		})

		Convey("each call produces unique values", func() {
			raw2, hash2, err2 := GenerateSecureToken()
			So(err2, ShouldBeNil)
			So(raw, ShouldNotEqual, raw2)
			So(hash, ShouldNotEqual, hash2)
		})
	})
}

func TestGenerateJTI(t *testing.T) {
	Convey("generateJTI", t, func() {
		jti, err := generateJTI()

		Convey("returns no error", func() {
			So(err, ShouldBeNil)
		})

		Convey("returns 64-char hex string (32 bytes)", func() {
			So(len(jti), ShouldEqual, 64)
		})

		Convey("each call produces unique value", func() {
			jti2, err2 := generateJTI()
			So(err2, ShouldBeNil)
			So(jti, ShouldNotEqual, jti2)
		})
	})
}

func TestGenerateAccessToken(t *testing.T) {
	Convey("GenerateAccessToken", t, func() {
		tkns := newTokens("secret", "", "issuer", "audience")
		user := &authdomain.User{ID: "user-123", Role: authdomain.RoleAdmin}

		token, err := tkns.GenerateAccessToken(user, 15*time.Minute)

		Convey("returns no error", func() {
			So(err, ShouldBeNil)
		})

		Convey("token is a 3-part JWT", func() {
			So(len(strings.Split(token, ".")), ShouldEqual, 3)
		})

		Convey("claims contain correct fields", func() {
			claims, err2 := tkns.ValidateToken(token)
			So(err2, ShouldBeNil)
			So(claims.Subject, ShouldEqual, "user-123")
			So(claims.Role, ShouldEqual, string(authdomain.RoleAdmin))
			So(claims.Issuer, ShouldEqual, "issuer")
			So(claims.Audience, ShouldResemble, jwt.ClaimStrings{"audience"})
			So(claims.ID, ShouldNotBeEmpty)
			So(claims.IssuedAt, ShouldNotBeNil)
			So(claims.ExpiresAt, ShouldNotBeNil)
		})
	})
}

func TestValidateToken(t *testing.T) {
	Convey("ValidateToken", t, func() {
		const (
			secret   = "current-secret"
			prevSec  = "prev-secret"
			issuer   = "learnflow"
			audience = "web"
		)
		user := &authdomain.User{ID: "user-123", Role: authdomain.RoleUser}

		Convey("valid token validates successfully", func() {
			tkns := newTokens(secret, "", issuer, audience)
			token, err := tkns.GenerateAccessToken(user, 15*time.Minute)
			So(err, ShouldBeNil)
			claims, err := tkns.ValidateToken(token)
			So(err, ShouldBeNil)
			So(claims.Subject, ShouldEqual, "user-123")
			So(claims.Role, ShouldEqual, string(authdomain.RoleUser))
		})

		Convey("expired token returns error", func() {
			tkns := newTokens(secret, "", issuer, audience)
			token, err := tkns.GenerateAccessToken(user, -time.Minute)
			So(err, ShouldBeNil)
			_, err = tkns.ValidateToken(token)
			So(err, ShouldNotBeNil)
		})

		Convey("token signed with wrong secret returns error", func() {
			signer := newTokens("other-secret", "", issuer, audience)
			token, err := signer.GenerateAccessToken(user, 15*time.Minute)
			So(err, ShouldBeNil)

			tkns := newTokens(secret, "", issuer, audience)
			_, err = tkns.ValidateToken(token)
			So(err, ShouldNotBeNil)
		})

		Convey("key rotation: token signed with prevSecret validates", func() {
			old := newTokens(prevSec, "", issuer, audience)
			token, err := old.GenerateAccessToken(user, 15*time.Minute)
			So(err, ShouldBeNil)

			rotated := newTokens(secret, prevSec, issuer, audience)
			claims, err := rotated.ValidateToken(token)
			So(err, ShouldBeNil)
			So(claims.Subject, ShouldEqual, "user-123")
		})

		Convey("malformed token returns error", func() {
			tkns := newTokens(secret, "", issuer, audience)
			_, err := tkns.ValidateToken("not.a.valid.token")
			So(err, ShouldNotBeNil)
		})

		Convey("token with wrong audience returns error", func() {
			tkns := newTokens(secret, "", issuer, audience)
			token, err := tkns.GenerateAccessToken(user, 15*time.Minute)
			So(err, ShouldBeNil)

			wrongAud := newTokens(secret, "", issuer, "other-audience")
			_, err = wrongAud.ValidateToken(token)
			So(err, ShouldNotBeNil)
		})

		Convey("token with wrong issuer returns error", func() {
			tkns := newTokens(secret, "", issuer, audience)
			token, err := tkns.GenerateAccessToken(user, 15*time.Minute)
			So(err, ShouldBeNil)

			wrongIss := newTokens(secret, "", "other-issuer", audience)
			_, err = wrongIss.ValidateToken(token)
			So(err, ShouldNotBeNil)
		})
	})
}
