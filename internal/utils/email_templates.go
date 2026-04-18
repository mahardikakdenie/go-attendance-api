package utils

import (
	"fmt"
	"strings"
)

func GetWelcomeEmailTemplate(name, email, password, companyName, logoURL string) string {
	// Branding Logic: If no logo, use initials
	brandingContent := ""
	if logoURL != "" {
		brandingContent = fmt.Sprintf(`<img src="%s" alt="%s Logo" style="max-height: 80px; margin-bottom: 20px;">`, logoURL, companyName)
	} else {
		initials := ""
		parts := strings.Split(companyName, " ")
		for i, part := range parts {
			if i < 2 && len(part) > 0 {
				initials += strings.ToUpper(string(part[0]))
			}
		}
		brandingContent = fmt.Sprintf(`
			<div style="width: 80px; height: 80px; background-color: #61AFEF; color: white; border-radius: 50%%; display: inline-flex; align-items: center; justify-content: center; font-size: 32px; font-weight: bold; margin: 0 auto 20px auto; line-height: 80px;">
				%s
			</div>`, initials)
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            margin: 0;
            padding: 0;
            background-color: #f4f7f9;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            background: #ffffff;
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 10px 25px rgba(0,0,0,0.05);
            border: 1px solid #eef2f5;
        }
        .header {
            background-color: #ffffff;
            padding: 40px 20px;
            text-align: center;
            border-bottom: 1px solid #f0f0f0;
        }
        .content {
            padding: 40px 35px;
        }
        .welcome-text {
            font-size: 22px;
            color: #1a1a1a;
            font-weight: 700;
            margin-bottom: 10px;
        }
        .company-badge {
            display: inline-block;
            background-color: #e3f2fd;
            color: #1976d2;
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 14px;
            font-weight: 600;
            margin-bottom: 25px;
        }
        .credentials {
            background-color: #fafbfc;
            border-radius: 10px;
            padding: 25px;
            margin: 30px 0;
            border: 1px dashed #d1d5db;
        }
        .credential-item {
            margin-bottom: 15px;
        }
        .credential-item:last-child {
            margin-bottom: 0;
        }
        .label {
            font-size: 13px;
            color: #6b7280;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 5px;
            display: block;
        }
        .value {
            font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
            font-size: 16px;
            color: #111827;
            font-weight: 600;
        }
        .button-container {
            text-align: center;
            margin: 35px 0;
        }
        .button {
            background-color: #111827;
            color: #ffffff !important;
            padding: 16px 35px;
            text-decoration: none;
            border-radius: 8px;
            font-weight: 600;
            display: inline-block;
            font-size: 16px;
        }
        .footer {
            background-color: #f9fafb;
            color: #9ca3af;
            padding: 25px;
            text-align: center;
            font-size: 13px;
            border-top: 1px solid #f3f4f6;
        }
        .warning-box {
            background-color: #fffbeb;
            border: 1px solid #fef3c7;
            border-radius: 8px;
            padding: 15px;
            margin-top: 25px;
        }
        .warning-text {
            color: #92400e;
            font-size: 14px;
            margin: 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            %s
            <div class="company-badge">%s</div>
            <h1 class="welcome-text">Account Activation</h1>
        </div>
        <div class="content">
            <p style="margin-top:0;">Hello <strong>%s</strong>,</p>
            <p>Welcome aboard! Your workspace at <strong>%s</strong> is ready. We've created an account for you to access our Attendance & HR platform.</p>
            
            <div class="credentials">
                <div class="credential-item">
                    <span class="label">Email Address</span>
                    <span class="value">%s</span>
                </div>
                <div class="credential-item">
                    <span class="label">Temporary Password</span>
                    <span class="value">%s</span>
                </div>
            </div>

            <div class="warning-box">
                <p class="warning-text"><strong>Security Note:</strong> This is a temporary password. You will be required to create a new, secure password upon your first sign-in.</p>
            </div>

            <div class="button-container">
                <a href="#" class="button">Access Workspace</a>
            </div>

            <p style="font-size: 14px; color: #6b7280; text-align: center;">If you didn't expect this invitation, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>Powered by Attendance Management System</p>
            <p>&copy; 2026 %s. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, brandingContent, companyName, name, companyName, email, password, companyName)
}

func GetLeaveApprovalRequestTemplate(approverName, requesterName, leaveType, startDate, endDate string, totalDays int, reason string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            margin: 0;
            padding: 0;
            background-color: #f4f7f9;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            background: #ffffff;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .header {
            background: linear-gradient(135deg, #F59E0B 0%%, #D97706 100%%);
            color: #ffffff;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            letter-spacing: 0.5px;
        }
        .content {
            padding: 30px;
        }
        .greeting {
            font-size: 18px;
            color: #282C34;
            font-weight: bold;
        }
        .details-box {
            background-color: #fffbeb;
            border-left: 4px solid #F59E0B;
            padding: 20px;
            margin: 20px 0;
            border-radius: 0 8px 8px 0;
        }
        .details-box p {
            margin: 8px 0;
            font-size: 15px;
        }
        .label {
            font-weight: bold;
            color: #555;
            display: inline-block;
            width: 100px;
        }
        .button-container {
            text-align: center;
            margin: 30px 0;
        }
        .button {
            background-color: #F59E0B;
            color: #ffffff;
            padding: 12px 25px;
            text-decoration: none;
            border-radius: 5px;
            font-weight: bold;
            display: inline-block;
            transition: background-color 0.3s;
        }
        .footer {
            background-color: #f4f7f9;
            color: #777;
            padding: 20px;
            text-align: center;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Leave Approval Required</h1>
        </div>
        <div class="content">
            <p class="greeting">Hello %s,</p>
            <p>You have a new leave request that requires your approval.</p>
            
            <div class="details-box">
                <p><span class="label">Employee:</span> %s</p>
                <p><span class="label">Leave Type:</span> %s</p>
                <p><span class="label">Duration:</span> %d Days</p>
                <p><span class="label">Date:</span> %s to %s</p>
                <p><span class="label">Reason:</span> %s</p>
            </div>

            <div class="button-container">
                <a href="#" class="button">Review Request</a>
            </div>

            <p>Please log in to the HR dashboard to approve or reject this request.</p>
        </div>
        <div class="footer">
            <p>&copy; 2026 Attendance API. All rights reserved.</p>
            <p>This is an automated message, please do not reply.</p>
        </div>
    </div>
</body>
</html>
`, approverName, requesterName, leaveType, totalDays, startDate, endDate, reason)
}

func GetLeaveDelegationTemplate(delegateName, requesterName, startDate, endDate string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            margin: 0;
            padding: 0;
            background-color: #f4f7f9;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            background: #ffffff;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .header {
            background: linear-gradient(135deg, #8B5CF6 0%%, #6D28D9 100%%);
            color: #ffffff;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
            letter-spacing: 0.5px;
        }
        .content {
            padding: 30px;
        }
        .greeting {
            font-size: 18px;
            color: #282C34;
            font-weight: bold;
        }
        .info-box {
            background-color: #f5f3ff;
            border-left: 4px solid #8B5CF6;
            padding: 20px;
            margin: 20px 0;
            border-radius: 0 8px 8px 0;
        }
        .info-box p {
            margin: 8px 0;
            font-size: 15px;
        }
        .highlight {
            font-weight: bold;
            color: #6D28D9;
        }
        .footer {
            background-color: #f4f7f9;
            color: #777;
            padding: 20px;
            text-align: center;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Task Delegation Notice</h1>
        </div>
        <div class="content">
            <p class="greeting">Hello %s,</p>
            <p>You have been assigned as a delegate for <span class="highlight">%s</span>.</p>
            
            <div class="info-box">
                <p><strong>Period:</strong> %s to %s</p>
                <p>During this period, any tasks, approvals, or responsibilities assigned to %s may be routed to you.</p>
            </div>

            <p>Please coordinate with them if you need handover details.</p>
        </div>
        <div class="footer">
            <p>&copy; 2026 Attendance API. All rights reserved.</p>
            <p>This is an automated message, please do not reply.</p>
        </div>
    </div>
</body>
</html>
`, delegateName, requesterName, startDate, endDate, requesterName)
}

func GetTrialConfirmationEmailTemplate(name, company string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            margin: 0;
            padding: 0;
            background-color: #f4f7f9;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            background: #ffffff;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .header {
            background: linear-gradient(135deg, #3B82F6 0%%, #1D4ED8 100%%);
            color: #ffffff;
            padding: 40px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 26px;
        }
        .content {
            padding: 30px;
        }
        .greeting {
            font-size: 18px;
            color: #1E3A8A;
            font-weight: bold;
        }
        .info-box {
            background-color: #EFF6FF;
            border-left: 4px solid #3B82F6;
            padding: 20px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .footer {
            background-color: #f4f7f9;
            color: #777;
            padding: 20px;
            text-align: center;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Trial Request Received</h1>
        </div>
        <div class="content">
            <p class="greeting">Hello %s,</p>
            <p>Thank you for your interest in our Attendance System for <strong>%s</strong>.</p>
            
            <p>Your trial request has been successfully received and is now being reviewed by our team. We will verify your information and get back to you with your account activation details shortly.</p>

            <div class="info-box">
                <p style="margin:0;"><strong>What's next?</strong></p>
                <p style="margin:5px 0 0 0;">You don't need to do anything. Simply wait for an email from us once your trial account is ready for use.</p>
            </div>

            <p>If you have any questions in the meantime, feel free to reply to this email.</p>
            
            <p>Best regards,<br><strong>Customer Success Team</strong></p>
        </div>
        <div class="footer">
            <p>&copy; 2026 Attendance API. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, name, company)
}

func GetCalendarEventReminderTemplate(userName, eventName, eventDate, eventType, description string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            margin: 0;
            padding: 0;
            background-color: #f4f7f9;
        }
        .container {
            max-width: 600px;
            margin: 20px auto;
            background: #ffffff;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 4px 15px rgba(0,0,0,0.1);
        }
        .header {
            background: linear-gradient(135deg, #10B981 0%%, #059669 100%%);
            color: #ffffff;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 24px;
        }
        .content {
            padding: 30px;
        }
        .greeting {
            font-size: 18px;
            color: #064E3B;
            font-weight: bold;
        }
        .event-box {
            background-color: #ECFDF5;
            border-left: 4px solid #10B981;
            padding: 20px;
            margin: 20px 0;
            border-radius: 0 4px 4px 0;
        }
        .event-box p {
            margin: 8px 0;
            font-size: 15px;
        }
        .label {
            font-weight: bold;
            color: #065F46;
            display: inline-block;
            width: 100px;
        }
        .footer {
            background-color: #f4f7f9;
            color: #777;
            padding: 20px;
            text-align: center;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Upcoming Event Reminder</h1>
        </div>
        <div class="content">
            <p class="greeting">Hello %s,</p>
            <p>This is a reminder for an upcoming event scheduled for tomorrow.</p>
            
            <div class="event-box">
                <p><span class="label">Event:</span> %s</p>
                <p><span class="label">Date:</span> %s</p>
                <p><span class="label">Type:</span> %s</p>
                <p><span class="label">Description:</span> %s</p>
            </div>

            <p>Please make sure to check your schedule and prepare accordingly.</p>
        </div>
        <div class="footer">
            <p>&copy; 2026 Attendance API. All rights reserved.</p>
            <p>This is an automated message, please do not reply.</p>
        </div>
    </div>
</body>
</html>
`, userName, eventName, eventDate, eventType, description)
}
