package utils

import "fmt"

func GetWelcomeEmailTemplate(name, email, password string) string {
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
            background: linear-gradient(135deg, #61AFEF 0%%, #282C34 100%%);
            color: #ffffff;
            padding: 40px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 28px;
            letter-spacing: 1px;
        }
        .content {
            padding: 30px;
        }
        .welcome-text {
            font-size: 18px;
            color: #282C34;
            font-weight: bold;
        }
        .credentials {
            background-color: #f8f9fa;
            border-left: 4px solid #61AFEF;
            padding: 20px;
            margin: 20px 0;
        }
        .credentials p {
            margin: 5px 0;
            font-family: 'Courier New', Courier, monospace;
        }
        .button-container {
            text-align: center;
            margin: 30px 0;
        }
        .button {
            background-color: #61AFEF;
            color: #ffffff;
            padding: 15px 30px;
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
        .warning {
            color: #E06C75;
            font-weight: bold;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to Attendance System</h1>
        </div>
        <div class="content">
            <p class="welcome-text">Hello, %s!</p>
            <p>Your account has been successfully created. You can now log in to the system using the following credentials:</p>
            
            <div class="credentials">
                <p><strong>Email:</strong> %s</p>
                <p><strong>Temporary Password:</strong> %s</p>
            </div>

            <p class="warning">⚠️ IMPORTANT: For your security, please log in and update your password immediately after your first access.</p>

            <div class="button-container">
                <a href="#" class="button">Login to Your Account</a>
            </div>

            <p>If you have any questions or encounter any issues, please contact your administrator.</p>
        </div>
        <div class="footer">
            <p>&copy; 2026 Attendance API. All rights reserved.</p>
            <p>This is an automated message, please do not reply.</p>
        </div>
    </div>
</body>
</html>
`, name, email, password)
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
