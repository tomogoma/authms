package sms_test

import (
	"github.com/tomogoma/authms/auth/phone/sms"
	"net/http"
	"net/http/httptest"
	"testing"
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
		desc   string
		conf   sms.ATConfig
		opts   []sms.ATOption
		expErr bool
	}{
		{
			desc:   "successful normal use case",
			conf:   &ATConfigMock{ExpUserName: "some-name", ExpAPIKey: "some-api-key"},
			opts:   make([]sms.ATOption, 0),
			expErr: false,
		},
		{
			desc:   "successful with URL",
			conf:   &ATConfigMock{ExpUserName: "some-name", ExpAPIKey: "some-api-key"},
			opts:   []sms.ATOption{sms.ATWithSendURL("http://some.valid.url")},
			expErr: false,
		},
		{
			desc:   "nil config",
			conf:   nil,
			opts:   make([]sms.ATOption, 0),
			expErr: true,
		},
		{
			desc:   "missing username",
			conf:   &ATConfigMock{ExpAPIKey: "some-api-key"},
			opts:   make([]sms.ATOption, 0),
			expErr: true,
		},
		{
			desc:   "missing API key",
			conf:   &ATConfigMock{ExpUserName: "some-name"},
			opts:   make([]sms.ATOption, 0),
			expErr: true,
		},
		{
			desc:   "empty URL ATOption",
			conf:   &ATConfigMock{ExpUserName: "some-name", ExpAPIKey: "some-api-key"},
			opts:   []sms.ATOption{sms.ATWithSendURL("")},
			expErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			at, err := sms.NewAfricasTalking(tc.conf, tc.opts...)
			if tc.expErr {
				if err == nil {
					t.Fatalf("Expected an error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("sms.NewAfricasTalking(): %v", err)
			}
			if at == nil {
				t.Fatalf("Received a nil AfricasTalking struct!")
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
	conf := &ATConfigMock{ExpUserName: "username", ExpAPIKey: "api-key"}
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
			at, err := sms.NewAfricasTalking(conf, sms.ATWithSendURL(ts.URL))
			if err != nil {
				t.Fatalf("sms.NewAfricasTalking(): %v", err)
			}
			err = at.SMS(toPhone, smsContent)
			if recUserName != conf.ExpUserName {
				t.Errorf("Submitted Username mismatch. Expect '%s', got '%s'",
					conf.ExpUserName, recUserName)
			}
			if recAPIKey != conf.ExpAPIKey {
				t.Errorf("Submitted API key mismatch. Expect '%s', got '%s'",
					conf.ExpAPIKey, recAPIKey)
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
				t.Fatalf("sms.AfricasTalking#SMS(): %v", err)
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
		{desc: "empty smsContent", toPhone: "0712345678", smsContent: ""},
	}
	conf := &ATConfigMock{ExpUserName: "username", ExpAPIKey: "api-key"}
	at, err := sms.NewAfricasTalking(conf)
	if err != nil {
		t.Fatalf("sms.NewAfricasTalking(): %v", err)
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
