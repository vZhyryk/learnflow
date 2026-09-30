package events

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsKnownEventType(t *testing.T) {
	Convey("IsKnownEventType", t, func() {
		Convey("returns true for known event types", func() {
			So(IsKnownEventType(EventUserRegistered), ShouldBeTrue)
			So(IsKnownEventType(EventEmailChange), ShouldBeTrue)
			So(IsKnownEventType(EventAccountRecovery), ShouldBeTrue)
			So(IsKnownEventType(EventPasswordReset), ShouldBeTrue)
			So(IsKnownEventType(EventBriefSubmitted), ShouldBeTrue)
			So(IsKnownEventType(EventBookingCreated), ShouldBeTrue)
			So(IsKnownEventType(EventPaymentCompleted), ShouldBeTrue)
			So(IsKnownEventType(EventNotificationSend), ShouldBeTrue)
			So(IsKnownEventType(EventRegistrationAttemptOnExistingEmail), ShouldBeTrue)
		})

		Convey("returns false for unknown event types", func() {
			So(IsKnownEventType("unknown.event"), ShouldBeFalse)
		})
	})
}

func TestIsKnownAggregationType(t *testing.T) {
	Convey("IsKnownAggregationType", t, func() {
		Convey("returns true for known event types", func() {
			So(IsKnownAggregationType(AggregationTypeUser), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypeEmail), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypeAccount), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypePassword), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypeBrief), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypeBooking), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypePayment), ShouldBeTrue)
			So(IsKnownAggregationType(AggregationTypeAnnouncement), ShouldBeTrue)
		})

		Convey("returns false for unknown event types", func() {
			So(IsKnownAggregationType("unknown.event"), ShouldBeFalse)
		})
	})
}

// eventTypeValue returns the string value of a const spec declared with the EventType type.
func eventTypeValue(spec ast.Spec) (EventType, bool) {
	value, isValue := spec.(*ast.ValueSpec)
	if !isValue || value.Type == nil || len(value.Values) != 1 {
		return "", false
	}

	typeIdent, isIdent := value.Type.(*ast.Ident)
	if !isIdent || typeIdent.Name != "EventType" {
		return "", false
	}

	lit, isLit := value.Values[0].(*ast.BasicLit)
	if !isLit {
		return "", false
	}

	unquoted, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}

	return EventType(unquoted), true
}

// declaredEventTypes parses types.go and returns every constant declared with the EventType type.
func declaredEventTypes(t *testing.T) []EventType {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), "types.go", nil, 0)
	if err != nil {
		t.Fatalf("parse types.go: %v", err)
	}

	var declared []EventType
	for _, decl := range file.Decls {
		gen, isGen := decl.(*ast.GenDecl)
		if !isGen || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			if eventType, ok := eventTypeValue(spec); ok {
				declared = append(declared, eventType)
			}
		}
	}

	return declared
}

func TestEveryDeclaredEventTypeIsKnown(t *testing.T) {
	Convey("Given every EventType constant declared in types.go", t, func() {
		declared := declaredEventTypes(t)

		So(len(declared), ShouldBeGreaterThan, 10)
		for _, eventType := range declared {
			Convey(string(eventType)+" is registered in IsKnownEventType", func() {
				So(IsKnownEventType(eventType), ShouldBeTrue)
			})
		}
	})
}

func TestTokenPayloadIdempotencyKeyHidesRawToken(t *testing.T) {
	Convey("TokenPayload.GetIdempotencyKey", t, func() {
		p := TokenPayload{UserID: "user-123", RawToken: "super-secret-token"}

		key := p.GetIdempotencyKey()

		So(key, ShouldHaveLength, 2)
		So(key[0], ShouldEqual, "user-123")
		So(key[1], ShouldHaveLength, 64)
		So(key[1], ShouldNotContainSubstring, "super-secret-token")
		So(TokenPayload{UserID: "user-123", RawToken: "super-secret-token"}.GetIdempotencyKey(), ShouldResemble, key)
		So(TokenPayload{UserID: "user-123", RawToken: "another-token"}.GetIdempotencyKey(), ShouldNotResemble, key)
	})
}
