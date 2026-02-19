package email

import (
	"bytes"
	"html/template"
)

type otpTemplateData struct {
	AppName string
	Title   string
	Intro   string
	OtpCode string
}

const otpHTMLTmpl = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.AppName}}</title>
</head>
<body style="margin:0;padding:0;background:#f5f5f5;font-family:Arial,Helvetica,sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background:#f5f5f5;padding:40px 0;">
    <tr>
      <td align="center">
        <table width="480" cellpadding="0" cellspacing="0"
               style="background:#ffffff;border-radius:10px;box-shadow:0 2px 12px rgba(0,0,0,.08);overflow:hidden;">
          <!-- Header -->
          <tr>
            <td style="background:#1a1a2e;padding:28px 40px;">
              <p style="margin:0;font-size:22px;font-weight:700;color:#ffffff;letter-spacing:1px;">{{.AppName}}</p>
            </td>
          </tr>
          <!-- Body -->
          <tr>
            <td style="padding:40px;">
              <h2 style="margin:0 0 16px;font-size:20px;color:#1a1a1a;">{{.Title}}</h2>
              <p style="margin:0 0 32px;font-size:15px;color:#555;line-height:1.7;">{{.Intro}}</p>
              <!-- OTP box -->
              <table width="100%" cellpadding="0" cellspacing="0">
                <tr>
                  <td align="center">
                    <div style="display:inline-block;background:#f0f4ff;border:2px dashed #4a6cf7;border-radius:10px;padding:20px 40px;margin-bottom:32px;">
                      <span style="font-size:40px;font-weight:900;letter-spacing:16px;color:#1a1a2e;font-family:'Courier New',Courier,monospace;">{{.OtpCode}}</span>
                    </div>
                  </td>
                </tr>
              </table>
              <p style="margin:0 0 8px;font-size:14px;color:#777;line-height:1.6;">
                Kode ini berlaku selama <strong>10 menit</strong>.
              </p>
              <p style="margin:0;font-size:14px;color:#e74c3c;line-height:1.6;">
                Jangan bagikan kode ini kepada siapapun, termasuk tim {{.AppName}}.
              </p>
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td style="background:#f9f9f9;padding:20px 40px;border-top:1px solid #eee;">
              <p style="margin:0;font-size:12px;color:#aaa;text-align:center;">
                Email ini dikirim otomatis oleh sistem {{.AppName}}.
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

var parsedOTPTemplate = template.Must(template.New("otp").Parse(otpHTMLTmpl))

type otpPurposeInfo struct {
	title string
	intro string
}

var purposeLabels = map[string]otpPurposeInfo{
	"email_verification": {
		title: "Verifikasi Email Kamu",
		intro: "Terima kasih telah mendaftar! Gunakan kode berikut untuk memverifikasi alamat email kamu:",
	},
	"reset_password": {
		title: "Reset Password",
		intro: "Kami menerima permintaan reset password untuk akunmu. Gunakan kode berikut untuk melanjutkan:",
	},
	"login": {
		title: "Kode Login OTP",
		intro: "Gunakan kode berikut untuk masuk ke akun kamu:",
	},
}

// BuildOTPEmailHTML returns rendered HTML for the given OTP purpose and code.
func BuildOTPEmailHTML(appName, otpCode, purpose string) (string, error) {
	info, ok := purposeLabels[purpose]
	if !ok {
		info = otpPurposeInfo{title: "Kode OTP", intro: "Gunakan kode berikut:"}
	}
	data := otpTemplateData{
		AppName: appName,
		Title:   info.title,
		Intro:   info.intro,
		OtpCode: otpCode,
	}
	var buf bytes.Buffer
	if err := parsedOTPTemplate.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
