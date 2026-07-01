package authdomain

import (
	"errors"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestErrAccountLockedError(t *testing.T) {
	Convey("ErrAccountLockedError", t, func() {
		Convey("Error returns account locked message", func() {
			e := &ErrAccountLockedError{}
			So(e.Error(), ShouldEqual, "account locked")
		})
		Convey("Is matches ErrAccountLocked sentinel", func() {
			e := &ErrAccountLockedError{}
			So(errors.Is(e, ErrAccountLocked), ShouldBeTrue)
		})
		Convey("Is does not match unrelated error", func() {
			e := &ErrAccountLockedError{}
			So(errors.Is(e, ErrInvalidToken), ShouldBeFalse)
		})
	})
}

func TestReasonValid(t *testing.T) {
	Convey("Reason validation", t, func() {
		Convey("Is Valid", func() {
			So(RevokeReasonLogout.Valid(), ShouldBeTrue)
		})
		Convey("is invalid", func() {
			reason := RevokeReason("invalid")
			So(reason.Valid(), ShouldBeFalse)
		})
	})
}

func TestRegisterRequestValidate(t *testing.T) {
	Convey("Register request validation", t, func() {
		Convey("nil is not allowed", func() {
			err := (&RegisterRequest{}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email string is empty", func() {
			email := ""
			err := (&RegisterRequest{Email: email}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email too short", func() {
			email := strings.Repeat("a", 2)
			err := (&RegisterRequest{Email: email}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("invalid email format", func() {
			email := strings.Repeat("a", 5)
			err := (&RegisterRequest{Email: email}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("password string is empty", func() {
			email := "test@gmail.com"
			password := ""
			err := (&RegisterRequest{Email: email, Password: password}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("password too short", func() {
			email := "test@gmail.com"
			password := strings.Repeat("a", 5)
			err := (&RegisterRequest{Email: email, Password: password}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("password too long", func() {
			email := "test@gmail.com"
			password := strings.Repeat("a", 73)
			err := (&RegisterRequest{Email: email, Password: password}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("valid request", func() {
			email := "test@gmail.com"
			password := "validpassword"
			So((&RegisterRequest{Email: email, Password: password}).Validate(), ShouldBeNil)
		})
	})
}

func TestLoginRequestValidate(t *testing.T) {
	Convey("Login request validation", t, func() {
		Convey("nil is not allowed", func() {
			err := (&LoginRequest{}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email string is empty", func() {
			err := (&LoginRequest{Email: "", Password: "validpass", UserAgent: "Mozilla/5.0"}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("invalid email format", func() {
			err := (&LoginRequest{Email: strings.Repeat("a", 5), Password: "validpass", UserAgent: "Mozilla/5.0"}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("password too short", func() {
			err := (&LoginRequest{Email: "test@gmail.com", Password: strings.Repeat("a", 5), UserAgent: "Mozilla/5.0"}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("password too long", func() {
			err := (&LoginRequest{Email: "test@gmail.com", Password: strings.Repeat("a", 73), UserAgent: "Mozilla/5.0"}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("user agent is empty", func() {
			err := (&LoginRequest{Email: "test@gmail.com", Password: "validpass", UserAgent: ""}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("user agent too long", func() {
			err := (&LoginRequest{Email: "test@gmail.com", Password: "validpass", UserAgent: strings.Repeat("a", 2001)}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("valid request", func() {
			So((&LoginRequest{Email: "test@gmail.com", Password: "validpass", UserAgent: "Mozilla/5.0"}).Validate(), ShouldBeNil)
		})
	})
}

func TestRefreshRequestValidate(t *testing.T) {
	Convey("Refresh request validation", t, func() {
		Convey("nil is not allowed", func() {
			err := (&RefreshRequest{}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email string is empty", func() {
			token := ""
			err := (&RefreshRequest{RefreshToken: token}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("valid request", func() {
			token := "valid-refresh-token"
			So((&RefreshRequest{RefreshToken: token}).Validate(), ShouldBeNil)
		})
	})
}

func TestLogoutRequestValidate(t *testing.T) {
	Convey("Logout request validation", t, func() {
		Convey("nil is not allowed", func() {
			err := (&LogoutRequest{}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email string is empty", func() {
			token := ""
			err := (&LogoutRequest{RefreshToken: token}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("valid request", func() {
			token := "valid-refresh-token"
			So((&LogoutRequest{RefreshToken: token}).Validate(), ShouldBeNil)
		})
	})
}

func TestVerifyEmailRequestValidate(t *testing.T) {
	Convey("Verify email request validation", t, func() {
		Convey("nil is not allowed", func() {
			err := (&VerifyEmailRequest{}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email string is empty", func() {
			token := ""
			err := (&VerifyEmailRequest{Token: token}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("valid request", func() {
			token := "valid-refresh-token"
			So((&VerifyEmailRequest{Token: token}).Validate(), ShouldBeNil)
		})
	})
}

func TestRequestPasswordResetRequestValidate(t *testing.T) {
	Convey("Request password reset request validation", t, func() {
		Convey("nil is not allowed", func() {
			err := (&RequestPasswordResetRequest{}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email string is empty", func() {
			email := ""
			err := (&RequestPasswordResetRequest{Email: email}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email too short", func() {
			email := strings.Repeat("a", 2)
			err := (&RequestPasswordResetRequest{Email: email}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("invalid email format", func() {
			email := strings.Repeat("a", 5)
			err := (&RequestPasswordResetRequest{Email: email}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("valid request", func() {
			email := "test@gmail.com"
			So((&RequestPasswordResetRequest{Email: email}).Validate(), ShouldBeNil)
		})
	})
}

func TestResetPasswordRequestValidate(t *testing.T) {
	Convey("Reset Password request validation", t, func() {
		Convey("nil is not allowed", func() {
			err := (&ResetPasswordRequest{}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("token string is empty", func() {
			token := ""
			err := (&ResetPasswordRequest{Token: token}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("password string is empty", func() {
			token := "valid-token"
			password := ""
			err := (&ResetPasswordRequest{Token: token, NewPassword: password}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("password too short", func() {
			token := "valid-token"
			password := strings.Repeat("a", 5)
			err := (&ResetPasswordRequest{Token: token, NewPassword: password}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("password too long", func() {
			token := "valid-token"
			password := strings.Repeat("a", 73)
			err := (&ResetPasswordRequest{Token: token, NewPassword: password}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("valid request", func() {
			token := "valid-token"
			password := "validpassword"
			So((&ResetPasswordRequest{Token: token, NewPassword: password}).Validate(), ShouldBeNil)
		})
	})
}

func TestChangePasswordRequestValidate(t *testing.T) {
	Convey("Change Password request validation", t, func() {
		Convey("nil is not allowed", func() {
			err := (&ChangePasswordRequest{}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("old password is empty", func() {
			oldPassword := ""
			err := (&ChangePasswordRequest{OldPassword: oldPassword}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("password string is empty", func() {
			oldPassword := "valid-oldPassword"
			password := ""
			err := (&ChangePasswordRequest{OldPassword: oldPassword, NewPassword: password}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("password too short", func() {
			oldPassword := "valid-oldPassword"
			password := strings.Repeat("a", 5)
			err := (&ChangePasswordRequest{OldPassword: oldPassword, NewPassword: password}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("password too long", func() {
			oldPassword := "valid-oldPassword"
			password := strings.Repeat("a", 73)
			err := (&ChangePasswordRequest{OldPassword: oldPassword, NewPassword: password}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("Same old and new passwords", func() {
			oldPassword := "valid-oldPassword"
			err := (&ChangePasswordRequest{OldPassword: oldPassword, NewPassword: oldPassword}).Validate()
			So(errors.Is(err, ErrWrongPassword), ShouldBeTrue)
		})

		Convey("valid request", func() {
			oldPassword := "valid-oldPassword"
			password := "valid-newPassword"
			So((&ChangePasswordRequest{OldPassword: oldPassword, NewPassword: password}).Validate(), ShouldBeNil)
		})
	})
}

func TestRequestEmailChangeRequestValidate(t *testing.T) {
	Convey("Request Email Change request validation", t, func() {
		Convey("nil is not allowed", func() {
			err := (&RequestEmailChangeRequest{}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email string is empty", func() {
			email := ""
			err := (&RequestEmailChangeRequest{NewEmail: email}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email too short", func() {
			email := strings.Repeat("a", 2)
			err := (&RequestEmailChangeRequest{NewEmail: email}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("invalid email format", func() {
			email := strings.Repeat("a", 5)
			err := (&RequestEmailChangeRequest{NewEmail: email}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("valid request", func() {
			email := "test@gmail.com"
			So((&RequestEmailChangeRequest{NewEmail: email}).Validate(), ShouldBeNil)
		})
	})
}

func TestEmailChangeRequestValidate(t *testing.T) {
	Convey("Email change request validation", t, func() {
		Convey("nil is not allowed", func() {
			err := (&EmailChangeRequest{}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email string is empty", func() {
			token := ""
			err := (&EmailChangeRequest{Token: token}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("valid request", func() {
			token := "valid-refresh-token"
			So((&EmailChangeRequest{Token: token}).Validate(), ShouldBeNil)
		})
	})
}

func TestRecoverAccountRequestValidate(t *testing.T) {
	Convey("Recover account request validation", t, func() {
		Convey("nil is not allowed", func() {
			err := (&RecoverAccountRequest{}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email string is empty", func() {
			token := ""
			err := (&RecoverAccountRequest{Token: token}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("valid request", func() {
			token := "valid-refresh-token"
			So((&RecoverAccountRequest{Token: token}).Validate(), ShouldBeNil)
		})
	})
}

func TestRequestRecoverAccountRequestValidate(t *testing.T) {
	Convey("Request Recover Account request validation", t, func() {
		Convey("nil is not allowed", func() {
			err := (&RequestRecoverAccountRequest{}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email string is empty", func() {
			email := ""
			err := (&RequestRecoverAccountRequest{Email: email}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})
		Convey("email too short", func() {
			email := strings.Repeat("a", 2)
			err := (&RequestRecoverAccountRequest{Email: email}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("invalid email format", func() {
			email := strings.Repeat("a", 5)
			err := (&RequestRecoverAccountRequest{Email: email}).Validate()
			So(errors.Is(err, ErrInvalidCredentialFormat), ShouldBeTrue)
		})

		Convey("valid request", func() {
			email := "test@gmail.com"
			So((&RequestRecoverAccountRequest{Email: email}).Validate(), ShouldBeNil)
		})
	})
}
