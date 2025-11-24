# MailHog Email Inbox Access

## Overview

MailHog is now accessible through nginx at `/inbox` for convenient email testing without needing to remember port numbers.

## Access URLs

### Localhost Development
```
http://localhost/inbox
https://localhost/inbox  (with HTTPS setup)
```

### Production/Staging
```
https://rex.stage.fauda.dream11.in/inbox
```

## What is MailHog?

MailHog is an email testing tool that captures all outgoing emails from your application during development and staging. Instead of sending real emails, they are captured and can be viewed in a web interface.

## Features

- **View all test emails** sent by the application
- **Real-time updates** via WebSocket
- **Search and filter** emails
- **View email HTML** and plain text
- **Download attachments**
- **No real emails sent** to actual users

## Usage Examples

### Testing User Registration Emails

1. Sign up a new user at `https://localhost/auth`
2. Open MailHog inbox at `https://localhost/inbox`
3. See the verification email that was "sent"
4. Click the verification link to test the flow

### Testing Invitation Emails

1. Create a tenant
2. Invite a user to the tenant
3. Check `/inbox` to see the invitation email
4. Copy the invitation link from the email
5. Test the invitation acceptance flow

### Testing Password Reset

1. Request password reset
2. Check `/inbox` for reset email
3. Use the reset link from the email
4. Test the password reset flow

## Technical Details

### Nginx Configuration

```nginx
# Upstream definition
upstream mailhog {
    server mailhog:8025;
}

# Location block
location /inbox {
    proxy_pass http://mailhog;
    proxy_http_version 1.1;
    
    # WebSocket support for real-time updates
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    
    # Standard proxy headers
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

### Docker Compose

MailHog service configuration:

```yaml
mailhog:
  image: nfqlt/mailhog:arm64
  container_name: utm-mailhog
  ports:
    - "1025:1025"  # SMTP port (internal)
    - "8025:8025"  # Web UI (now proxied via nginx)
```

### Application Configuration

Your Go application sends emails to MailHog's SMTP server:

```go
// SMTP Configuration
SMTP_HOST=mailhog
SMTP_PORT=1025
SMTP_FROM=noreply@example.com
```

## Features Available in UI

### Main View
- **Inbox list** - All captured emails
- **Search** - Filter by recipient, subject, etc.
- **Delete** - Remove test emails
- **Clear all** - Start fresh

### Email Details
- **From/To/Subject** - Email metadata
- **Plain text view** - Text version
- **HTML view** - Rendered HTML email
- **Source view** - Raw email source
- **Download** - Save email as .eml file

### Real-time Updates
- New emails appear automatically
- No page refresh needed
- WebSocket connection for instant updates

## Common Use Cases

### 1. Testing Email Templates

```go
// Send test email
func TestEmailTemplate(t *testing.T) {
    // Send email via your service
    emailService.SendWelcomeEmail("test@example.com", "John Doe")
    
    // Check /inbox manually or via API
    // Verify the email looks correct
}
```

### 2. Testing Email Links

1. Send invitation/verification email
2. Open `/inbox`
3. Copy the link from the email
4. Test the link in a new tab
5. Verify the flow works end-to-end

### 3. Testing Email Content

- Check subject lines
- Verify sender address
- Test email formatting
- Verify all links work
- Test with different email clients (MailHog renders HTML)

## API Access (Optional)

MailHog also has a JSON API if you want to automate email testing:

### List all messages
```bash
curl https://localhost/inbox/api/v2/messages | jq .
```

### Get specific message
```bash
curl https://localhost/inbox/api/v2/messages/MESSAGE_ID | jq .
```

### Delete all messages
```bash
curl -X DELETE https://localhost/inbox/api/v1/messages
```

## Troubleshooting

### Issue: Can't access /inbox

**Check nginx is running:**
```bash
docker-compose ps nginx
```

**Check MailHog is running:**
```bash
docker-compose ps mailhog
```

**Restart if needed:**
```bash
docker-compose restart nginx mailhog
```

### Issue: Emails not appearing

**Check SMTP configuration:**
```bash
# In your Go application
SMTP_HOST=mailhog
SMTP_PORT=1025  # Not 8025!
```

**Check MailHog logs:**
```bash
docker-compose logs mailhog --tail=50
```

### Issue: WebSocket connection failed

**Check nginx configuration:**
```bash
docker-compose exec nginx nginx -t
```

**Check nginx logs:**
```bash
docker-compose logs nginx | grep inbox
```

## Security Notes

### Development/Staging Only

⚠️ **MailHog should NEVER be used in production!**

- It captures ALL emails
- No authentication required
- Anyone can view all emails
- It's a testing tool only

### Production Email Setup

For production, use a real email service:
- AWS SES
- SendGrid
- Mailgun
- Postmark

Update your Go application configuration:
```go
// Production SMTP settings
SMTP_HOST=email-smtp.us-east-1.amazonaws.com
SMTP_PORT=587
SMTP_USER=your-aws-ses-user
SMTP_PASSWORD=your-aws-ses-password
```

## Alternative Access Methods

If you prefer the old way:

### Direct Port Access (Port 8025)
```
http://localhost:8025
```

### API Documentation
MailHog API docs: https://github.com/mailhog/MailHog/blob/master/docs/APIv2.md

## Related Documentation

- [Quickstart Guide](QUICKSTART.md) - Initial setup
- [Docker Compose](../docker-compose.yml) - Service configuration
- [Nginx Configuration](../nginx.conf) - Reverse proxy setup

---

**Last Updated**: November 24, 2025  
**Access URL**: `/inbox`  
**Status**: ✅ Available in Development & Staging

