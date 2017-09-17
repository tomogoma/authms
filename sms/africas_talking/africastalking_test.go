package africas_talking_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tomogoma/authms/sms/africas_talking"
)

type ATConfigMock struct {
	ExpUserName string
	ExpAPIKey   string
}

func (a *ATConfigMock) Username() string {
	return a.ExpUserName
}
func (a *ATConfigMock) APIKey() string {
	return a.ExpAPIKey
}

func TestNewAfricasTalking(t *testing.T) {
	testCases := []struct {
		desc     string
		username string
		apiKey   string
		opts     []africas_talking.Option
		expErr   bool
	}{
		{
			desc:     "successful normal use case",
			username: "some-name",
			apiKey:   "some-api-key",
			opts:     make([]africas_talking.Option, 0),
			expErr:   false,
		},
		{
			desc:     "successful with URL",
			username: "some-name",
			apiKey:   "some-api-key",
			opts:     []africas_talking.Option{africas_talking.SendURL("http://some.valid.url")},
			expErr:   false,
		},
		{
			desc:   "missing username",
			apiKey: "some-api-key",
			opts:   make([]africas_talking.Option, 0),
			expErr: true,
		},
		{
			desc:     "missing API key",
			username: "some-name",
			opts:     make([]africas_talking.Option, 0),
			expErr:   true,
		},
		{
			desc:     "empty URL Option",
			username: "some-name",
			apiKey:   "some-api-key",
			opts:     []africas_talking.Option{africas_talking.SendURL("")},
			expErr:   true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			at, err := africas_talking.NewSMSCl(tc.username, tc.apiKey, tc.opts...)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("africas_talking.NewSMSCl(): %v", err)
			}
			if at == nil {
				t.Fatalf("Received a nil SMSCl struct!")
			}
		})
	}
}

// a sample success response looks as follows:
//
//     <AfricasTalkingResponse>
//			<SMSMessageData>
//				<Message>Sent to 1/1 Total Cost: KES Y.YYYY</Message>
//				<Recipients>
//					<Recipient>
//						<number>+2547XXXXXXXX</number>
//						<cost>KES Y.YYYY</cost>
//						<status>Success</status>
//						<messageId>ATXid_2cc1431580607d58123e29bcd7b1d106</messageId>
//					</Recipient>
//				</Recipients>
//			</SMSMessageData>
//		</AfricasTalkingResponse>
//
// The value of interest is the status element which has the value "Success"
// on successful completion.
func TestAfricasTalking_SMS(t *testing.T) {
	testCases := []struct {
		desc      string
		expStatus int
		expBody   string
		expErr    bool
	}{
		{
			desc:      "successful request",
			expStatus: http.StatusOK,
			expBody: `
<AfricasTalkingResponse>
	<SMSMessageData>
		<Message>Sent to 1/1 Total Cost: KES Y.YYYY</Message>
		<Recipients>
			<Recipient>
				<number>+2547XXXXXXXX</number>
				<cost>KES Y.YYYY</cost>
				<status>Success</status>
				<messageId>ATXid_2cc1431580607d58123e29bcd7b1d106</messageId>
			</Recipient>
		</Recipients>
	</SMSMessageData>
</AfricasTalkingResponse>
			`,
			expErr: false,
		},
		{
			desc:      "http status error",
			expStatus: http.StatusRequestTimeout,
			expBody:   ``,
			expErr:    true,
		},
		{
			desc:      "none successful status in body",
			expStatus: http.StatusOK,
			expBody: `
<AfricasTalkingResponse>
	<SMSMessageData>
		<Message>Sent to 1/1 Total Cost: KES Y.YYYY</Message>
		<Recipients>
			<Recipient>
				<number>+2547XXXXXXXX</number>
				<cost>KES Y.YYYY</cost>
				<status>error - some error</status>
				<messageId>ATXid_2cc1431580607d58123e29bcd7b1d106</messageId>
			</Recipient>
		</Recipients>
	</SMSMessageData>
</AfricasTalkingResponse>
			`,
			expErr: true,
		},
		{
			desc:      "Unexpected results count",
			expStatus: http.StatusOK,
			expBody: `
<AfricasTalkingResponse>
	<SMSMessageData>
		<Message>Sent to 1/1 Total Cost: KES Y.YYYY</Message>
		<Recipients>
			<Recipient>
				<number>+2547XXXXXXXX</number>
				<cost>KES Y.YYYY</cost>
				<status>error - some error</status>
				<messageId>ATXid_2cc1431580607d58123e29bcd7b1d106</messageId>
			</Recipient>
			<Recipient>
				<number>+2547XXXXXXXX</number>
				<cost>KES Y.YYYY</cost>
				<status>error - some error</status>
				<messageId>ATXid_2cc1431580607d58123e29bcd7b1d106</messageId>
			</Recipient>
		</Recipients>
	</SMSMessageData>
</AfricasTalkingResponse>
			`,
			expErr: true,
		},
	}
	username := "username"
	APIKey := "api-key"
	toPhone := "+254712345678"
	smsContent := "This is a test SMS\nWith a new Line"
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			var recToPhone, recSMSContent, recUserName, recAPIKey string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				recUserName = q.Get("username")
				recAPIKey = q.Get("Apikey")
				recToPhone = q.Get("to")
				recSMSContent = q.Get("message")
				w.WriteHeader(tc.expStatus)
				w.Write([]byte(tc.expBody))
			}))
			defer ts.Close()
			at, err := africas_talking.NewSMSCl(username, APIKey, africas_talking.SendURL(ts.URL))
			if err != nil {
				t.Fatalf("africas_talking.NewSMSCl(): %v", err)
			}
			err = at.SMS(toPhone, smsContent)
			if recUserName != username {
				t.Errorf("Submitted Username mismatch. Expect '%s', got '%s'",
					username, recUserName)
			}
			if recAPIKey != APIKey {
				t.Errorf("Submitted API key mismatch. Expect '%s', got '%s'",
					APIKey, recAPIKey)
			}
			if recToPhone != toPhone {
				t.Errorf("Submitted to-phone mismatch. Expect '%s', got '%s'",
					toPhone, recToPhone)
			}
			if recSMSContent != smsContent {
				t.Errorf("Submitted SMS content mismatch. Expect '%s', got '%s'",
					smsContent, recSMSContent)
			}
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("africas_talking.SMSCl#SMS(): %v", err)
			}
		})
	}
}

func TestAfricasTalking_SMS_invalidParams(t *testing.T) {
	testCases := []struct {
		desc       string
		toPhone    string
		smsContent string
	}{
		{desc: "empty toPhone", toPhone: "", smsContent: "Some SMS content"},
		{desc: "empty sms Content", toPhone: "0712345678", smsContent: ""},
	}
	at, err := africas_talking.NewSMSCl("username", "api-key")
	if err != nil {
		t.Fatalf("africas_talking.NewSMSCl(): %v", err)
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := at.SMS(tc.toPhone, tc.smsContent)
			if err == nil {
				t.Fatalf("Expected an error but got nil")
			}
		})
	}
}
