# Error Codes Reference

Complete reference of all API error codes and their meanings.

## Error Response Format

All API errors follow this structure:

```json
{
  "success": false,
  "error": "Error message describing what went wrong"
}
```

## HTTP Status Codes

| Code | Name | Usage |
|------|------|-------|
| 200 | OK | Successful request |
| 201 | Created | Resource created successfully |
| 202 | Accepted | Request accepted for processing (async) |
| 204 | No Content | Successful deletion |
| 400 | Bad Request | Invalid request data |
| 401 | Unauthorized | Authentication required or failed |
| 403 | Forbidden | Permission denied |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource conflict (e.g., duplicate) |
| 422 | Unprocessable Entity | Validation failed |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server-side error |
| 503 | Service Unavailable | Service temporarily down |

## Authentication Errors (401)

### Missing Authentication

**Status**: `401 Unauthorized`

```json
{
  "success": false,
  "error": "Unauthorized: No session found"
}
```

**Cause**: No session cookie or Authorization header provided  
**Solution**: Include valid session cookie or System User token

### Invalid Token

**Status**: `401 Unauthorized`

```json
{
  "success": false,
  "error": "Invalid authentication token"
}
```

**Cause**: Token is malformed, expired, or invalid  
**Solution**: Obtain a new token

### Session Expired

**Status**: `401 Unauthorized`

```json
{
  "success": false,
  "error": "Session expired"
}
```

**Cause**: User session has expired  
**Solution**: Re-authenticate user

### System User Inactive

**Status**: `401 Unauthorized`

```json
{
  "success": false,
  "error": "System user is inactive"
}
```

**Cause**: System User credentials have been deactivated or expired  
**Solution**: Create new System User credentials

## Authorization Errors (403)

### Permission Denied

**Status**: `403 Forbidden`

```json
{
  "success": false,
  "error": "Permission denied"
}
```

**Cause**: User lacks required permission for this action  
**Solution**: Contact tenant admin to update your role/permissions

### Not Tenant Member

**Status**: `403 Forbidden`

```json
{
  "success": false,
  "error": "You are not a member of this tenant"
}
```

**Cause**: Attempting to access tenant you're not a member of  
**Solution**: Request invitation to the tenant

### Not Platform Admin

**Status**: `403 Forbidden`

```json
{
  "success": false,
  "error": "Platform admin access required"
}
```

**Cause**: Attempting platform admin operation without platform admin role  
**Solution**: Contact system administrator

### Tenant Suspended

**Status**: `403 Forbidden`

```json
{
  "success": false,
  "error": "Tenant is suspended"
}
```

**Cause**: Tenant has been suspended  
**Solution**: Contact platform administrator

## Validation Errors (400/422)

### Invalid Input

**Status**: `400 Bad Request`

```json
{
  "success": false,
  "error": "Invalid input: field 'name' is required"
}
```

**Common Causes**:
- Missing required fields
- Invalid field types
- Invalid field values

### Validation Failed

**Status**: `422 Unprocessable Entity`

```json
{
  "success": false,
  "error": "Validation failed: slug must be lowercase and alphanumeric"
}
```

**Common Validations**:
- **Tenant slug**: lowercase, alphanumeric, hyphens only
- **Email**: valid email format
- **UUID**: valid UUID format
- **Dates**: valid ISO 8601 format

### Invalid JSON

**Status**: `400 Bad Request`

```json
{
  "success": false,
  "error": "Invalid JSON in request body"
}
```

**Cause**: Malformed JSON in request body  
**Solution**: Validate JSON syntax

## Resource Not Found (404)

### Tenant Not Found

**Status**: `404 Not Found`

```json
{
  "success": false,
  "error": "Tenant not found"
}
```

**Causes**:
- Invalid tenant ID
- Tenant has been deleted
- No access to tenant

### User Not Found

**Status**: `404 Not Found`

```json
{
  "success": false,
  "error": "User not found"
}
```

### Member Not Found

**Status**: `404 Not Found`

```json
{
  "success": false,
  "error": "Member not found in this tenant"
}
```

### Role/Policy/Permission Not Found

**Status**: `404 Not Found`

```json
{
  "success": false,
  "error": "Role not found"
}
```

### System User Not Found

**Status**: `404 Not Found`

```json
{
  "success": false,
  "error": "System user not found"
}
```

### Invitation Not Found

**Status**: `404 Not Found`

```json
{
  "success": false,
  "error": "Invitation not found"
}
```

## Conflict Errors (409)

### Duplicate Tenant Slug

**Status**: `409 Conflict`

```json
{
  "success": false,
  "error": "Tenant slug already exists"
}
```

**Cause**: Slug is already in use  
**Solution**: Choose a different slug

### User Already Member

**Status**: `409 Conflict`

```json
{
  "success": false,
  "error": "User is already a member of this tenant"
}
```

**Cause**: Attempting to add user who is already a member  
**Solution**: Update existing membership instead

### Duplicate Invitation

**Status**: `409 Conflict`

```json
{
  "success": false,
  "error": "User already has a pending invitation"
}
```

**Cause**: User already has pending invitation to this tenant  
**Solution**: Resend or cancel existing invitation

### Duplicate Role Name

**Status**: `409 Conflict`

```json
{
  "success": false,
  "error": "Role with this name already exists"
}
```

### Duplicate Permission

**Status**: `409 Conflict`

```json
{
  "success": false,
  "error": "Permission already exists"
}
```

## Business Logic Errors (400)

### Cannot Delete Last Admin

**Status**: `400 Bad Request`

```json
{
  "success": false,
  "error": "Cannot remove the last admin member"
}
```

**Cause**: Attempting to remove last admin from tenant  
**Solution**: Assign admin role to another member first

### Cannot Delete System Role

**Status**: `400 Bad Request`

```json
{
  "success": false,
  "error": "Cannot delete system-defined role"
}
```

**Cause**: Attempting to delete built-in system role  
**Solution**: Only custom roles can be deleted

### Invitation Expired

**Status**: `400 Bad Request`

```json
{
  "success": false,
  "error": "Invitation has expired"
}
```

**Cause**: Attempting to accept expired invitation  
**Solution**: Request new invitation

### Invitation Already Accepted

**Status**: `400 Bad Request`

```json
{
  "success": false,
  "error": "Invitation has already been accepted"
}
```

## Server Errors (500)

### Database Error

**Status**: `500 Internal Server Error`

```json
{
  "success": false,
  "error": "Database error occurred"
}
```

**Cause**: Database connection or query failed  
**Action**: Contact support if persists

### External Service Error

**Status**: `500 Internal Server Error`

```json
{
  "success": false,
  "error": "Failed to initialize tenant in service"
}
```

**Cause**: External service unavailable or returned error  
**Action**: Check service status, retry operation

### Email Send Failed

**Status**: `500 Internal Server Error`

```json
{
  "success": false,
  "error": "Failed to send invitation email"
}
```

**Cause**: Email service unavailable or misconfigured  
**Action**: Check SMTP configuration

## Rate Limiting (429)

### Too Many Requests

**Status**: `429 Too Many Requests`

```json
{
  "success": false,
  "error": "Rate limit exceeded. Please try again later."
}
```

**Headers**:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1632150000
```

**Cause**: Exceeded API rate limit  
**Solution**: Wait until rate limit resets, implement backoff

## Service Unavailable (503)

### Maintenance Mode

**Status**: `503 Service Unavailable`

```json
{
  "success": false,
  "error": "Service temporarily unavailable for maintenance"
}
```

**Cause**: System is in maintenance mode  
**Action**: Wait for maintenance to complete

## Error Handling Best Practices

### Frontend

```typescript
async function apiCall() {
  try {
    const response = await fetch('/api/v1/tenants', {
      method: 'POST',
      credentials: 'include',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify(data)
    });
    
    if (!response.ok) {
      const error = await response.json();
      
      // Handle specific error codes
      switch (response.status) {
        case 401:
          // Redirect to login
          window.location.href = '/login';
          break;
        
        case 403:
          // Show permission denied message
          alert('Permission denied: ' + error.error);
          break;
        
        case 409:
          // Show conflict message
          alert('Conflict: ' + error.error);
          break;
        
        case 422:
          // Show validation errors
          displayValidationErrors(error.error);
          break;
        
        default:
          // Generic error
          alert('Error: ' + error.error);
      }
      
      return null;
    }
    
    return await response.json();
  } catch (err) {
    console.error('Network error:', err);
    alert('Network error. Please check your connection.');
    return null;
  }
}
```

### Backend (Go)

```go
// Custom error types
type APIError struct {
    StatusCode int
    Message    string
}

func (e *APIError) Error() string {
    return e.Message
}

// Error helpers
func NewBadRequestError(message string) *APIError {
    return &APIError{StatusCode: 400, Message: message}
}

func NewNotFoundError(message string) *APIError {
    return &APIError{StatusCode: 404, Message: message}
}

// Error handler middleware
func ErrorHandler(c *gin.Context) {
    c.Next()
    
    if len(c.Errors) > 0 {
        err := c.Errors.Last().Err
        
        if apiErr, ok := err.(*APIError); ok {
            c.JSON(apiErr.StatusCode, gin.H{
                "success": false,
                "error":   apiErr.Message,
            })
            return
        }
        
        // Generic error
        c.JSON(500, gin.H{
            "success": false,
            "error":   "Internal server error",
        })
    }
}
```

## Debugging Errors

### Enable Debug Mode

```bash
# In .env
APP_ENV=development
LOG_LEVEL=debug
```

### Check Logs

```bash
# API logs
docker logs utm-api | grep ERROR

# Worker logs
docker logs utm-worker | grep ERROR
```

### Common Issues

**"Session expired" immediately after login**:
- Check CORS configuration
- Verify cookie domain settings
- Ensure `credentials: 'include'` in fetch calls

**"Permission denied" for admin user**:
- Verify user's role assignment
- Check RBAC configuration
- Confirm permissions are assigned to policies

**"Tenant not found" for valid ID**:
- Check soft delete (`deleted_at` field)
- Verify tenant status
- Confirm user is a member

## Related Documentation

- [API Overview](/x-api/overview) - API structure
- [Troubleshooting](/troubleshooting/common-issues) - Common issues
- [Debug Mode](/troubleshooting/debug-mode) - Debugging guide
