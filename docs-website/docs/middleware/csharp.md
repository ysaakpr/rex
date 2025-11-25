# C# / .NET Middleware Library

.NET middleware for RBAC authorization.

## Overview

This library provides ASP.NET Core middleware for integrating RBAC authorization checks in .NET applications.

**Status**: 🚧 Planned - Not yet implemented

**When available**, this library will provide:
- ASP.NET Core middleware
- Authorization policies
- Permission checking attributes
- Tenant access verification
- Platform admin checks
- System user authentication

## Planned Features

### ASP.NET Core Middleware

```csharp
// Future implementation
using UtmRbac.AspNetCore;

var builder = WebApplication.CreateBuilder(args);

// Add RBAC services
builder.Services.AddUtmRbac(options =>
{
    options.ApiUrl = "http://localhost:8080";
    options.CacheEnabled = true;
    options.CacheTTL = TimeSpan.FromMinutes(5);
});

var app = builder.Build();

// Use RBAC middleware
app.UseUtmRbac();

app.MapPost("/api/posts", [RequirePermission("blog-api", "post", "create")]
async (CreatePostRequest request, HttpContext context) =>
{
    // Permission verified, user and tenant available
    var userId = context.User.GetUserId();
    var tenantId = context.User.GetTenantId();
    
    // Your handler logic
    return Results.Ok(new { Success = true });
});

app.Run();
```

### Authorization Attributes

```csharp
// Future implementation
using UtmRbac.Attributes;
using Microsoft.AspNetCore.Mvc;

[ApiController]
[Route("api/[controller]")]
public class PostsController : ControllerBase
{
    [HttpPost]
    [RequirePermission("blog-api", "post", "create")]
    public async Task<IActionResult> CreatePost([FromBody] CreatePostRequest request)
    {
        // Permission already verified
        var userId = User.GetUserId();
        var tenantId = User.GetTenantId();
        
        // Your logic
        return Ok(new { Success = true });
    }
    
    [HttpGet("{id}")]
    [RequireAnyPermission(
        new[] {
            new Permission("blog-api", "post", "read"),
            new Permission("blog-api", "post", "update")
        }
    )]
    public async Task<IActionResult> GetPost(string id)
    {
        // User has at least one of the permissions
        return Ok(GetPostData(id));
    }
    
    [HttpDelete("{id}")]
    [RequireAllPermissions(
        new[] {
            new Permission("blog-api", "post", "update"),
            new Permission("blog-api", "post", "delete")
        }
    )]
    public async Task<IActionResult> DeletePost(string id)
    {
        // User has all required permissions
        return Ok(new { Success = true });
    }
}
```

### Authorization Policies

```csharp
// Future implementation
using UtmRbac.Policies;

builder.Services.AddAuthorization(options =>
{
    // Define permission-based policies
    options.AddPermissionPolicy("CanCreatePost", 
        "blog-api", "post", "create");
    
    options.AddPermissionPolicy("CanPublishPost",
        "blog-api", "post", "publish");
    
    // Multiple permissions
    options.AddAnyPermissionPolicy("CanViewPost",
        new[] {
            new Permission("blog-api", "post", "read"),
            new Permission("blog-api", "post", "update")
        });
});

// Use in controller
[HttpPost]
[Authorize(Policy = "CanCreatePost")]
public async Task<IActionResult> CreatePost([FromBody] CreatePostRequest request)
{
    return Ok(new { Success = true });
}
```

## Current Workaround

Until the .NET library is available, use direct API calls:

```csharp
using System.Net.Http;
using System.Net.Http.Json;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Filters;

public class RequirePermissionAttribute : TypeFilterAttribute
{
    public RequirePermissionAttribute(string service, string entity, string action)
        : base(typeof(PermissionFilter))
    {
        Arguments = new object[] { service, entity, action };
    }
}

public class PermissionFilter : IAsyncActionFilter
{
    private readonly string _service;
    private readonly string _entity;
    private readonly string _action;
    private readonly IHttpClientFactory _httpClientFactory;
    
    public PermissionFilter(
        string service,
        string entity,
        string action,
        IHttpClientFactory httpClientFactory)
    {
        _service = service;
        _entity = entity;
        _action = action;
        _httpClientFactory = httpClientFactory;
    }
    
    public async Task OnActionExecutionAsync(
        ActionExecutingContext context,
        ActionExecutionDelegate next)
    {
        var userId = context.HttpContext.User.FindFirst("sub")?.Value;
        var tenantId = context.HttpContext.Request.RouteValues["tenantId"]?.ToString();
        
        if (string.IsNullOrEmpty(userId) || string.IsNullOrEmpty(tenantId))
        {
            context.Result = new UnauthorizedResult();
            return;
        }
        
        var client = _httpClientFactory.CreateClient();
        var url = $"http://localhost:8080/api/v1/authorize?" +
                  $"user_id={userId}&tenant_id={tenantId}&" +
                  $"service={_service}&entity={_entity}&action={_action}";
        
        try
        {
            var response = await client.GetFromJsonAsync<AuthorizeResponse>(url);
            
            if (response?.Data?.Authorized != true)
            {
                context.Result = new ForbidResult();
                return;
            }
            
            await next();
        }
        catch (Exception ex)
        {
            // Log error
            context.Result = new StatusCodeResult(500);
        }
    }
}

public class AuthorizeResponse
{
    public AuthorizeData Data { get; set; }
}

public class AuthorizeData
{
    public bool Authorized { get; set; }
}

// Usage in controller
[HttpPost]
[RequirePermission("blog-api", "post", "create")]
public async Task<IActionResult> CreatePost([FromBody] CreatePostRequest request)
{
    return Ok(new { Success = true });
}
```

### System User Client

```csharp
using System.Net.Http;
using System.Net.Http.Headers;
using System.Net.Http.Json;
using System.Text.Json;

public class SystemUserClient
{
    private readonly HttpClient _httpClient;
    
    public SystemUserClient(string token, string baseUrl = "http://localhost:8080")
    {
        _httpClient = new HttpClient
        {
            BaseAddress = new Uri(baseUrl)
        };
        
        _httpClient.DefaultRequestHeaders.Authorization = 
            new AuthenticationHeaderValue("Bearer", token);
        _httpClient.DefaultRequestHeaders.Accept.Add(
            new MediaTypeWithQualityHeaderValue("application/json"));
    }
    
    public async Task<List<Tenant>> GetTenantsAsync()
    {
        var response = await _httpClient.GetFromJsonAsync<TenantsResponse>(
            "/api/v1/tenants");
        return response?.Data?.Data ?? new List<Tenant>();
    }
    
    public async Task<List<Member>> GetMembersAsync(string tenantId)
    {
        var response = await _httpClient.GetFromJsonAsync<MembersResponse>(
            $"/api/v1/tenants/{tenantId}/members");
        return response?.Data?.Data ?? new List<Member>();
    }
    
    public async Task<Tenant> CreateTenantAsync(CreateTenantRequest request)
    {
        var response = await _httpClient.PostAsJsonAsync(
            "/api/v1/tenants", request);
        response.EnsureSuccessStatusCode();
        
        var result = await response.Content.ReadFromJsonAsync<TenantResponse>();
        return result?.Data;
    }
}

// Response models
public class TenantsResponse
{
    public TenantsData Data { get; set; }
}

public class TenantsData
{
    public List<Tenant> Data { get; set; }
}

public class Tenant
{
    public string Id { get; set; }
    public string Name { get; set; }
    public string Slug { get; set; }
}

// Usage
var token = Environment.GetEnvironmentVariable("SYSTEM_USER_TOKEN");
var client = new SystemUserClient(token);

var tenants = await client.GetTenantsAsync();
foreach (var tenant in tenants)
{
    var members = await client.GetMembersAsync(tenant.Id);
    Console.WriteLine($"{tenant.Name}: {members.Count} members");
}
```

### Reusable RBAC Service

```csharp
public interface IRbacService
{
    Task<bool> CheckPermissionAsync(
        string userId,
        string tenantId,
        string service,
        string entity,
        string action);
    
    Task<bool> CheckAnyPermissionAsync(
        string userId,
        string tenantId,
        IEnumerable<Permission> permissions);
    
    Task<List<Permission>> GetUserPermissionsAsync(
        string userId,
        string tenantId);
}

public class RbacService : IRbacService
{
    private readonly IHttpClientFactory _httpClientFactory;
    private readonly string _apiUrl;
    
    public RbacService(IHttpClientFactory httpClientFactory, IConfiguration config)
    {
        _httpClientFactory = httpClientFactory;
        _apiUrl = config["Rbac:ApiUrl"] ?? "http://localhost:8080";
    }
    
    public async Task<bool> CheckPermissionAsync(
        string userId,
        string tenantId,
        string service,
        string entity,
        string action)
    {
        var client = _httpClientFactory.CreateClient();
        var url = $"{_apiUrl}/api/v1/authorize?" +
                  $"user_id={userId}&tenant_id={tenantId}&" +
                  $"service={service}&entity={entity}&action={action}";
        
        try
        {
            var response = await client.GetFromJsonAsync<AuthorizeResponse>(url);
            return response?.Data?.Authorized ?? false;
        }
        catch
        {
            return false;
        }
    }
    
    public async Task<bool> CheckAnyPermissionAsync(
        string userId,
        string tenantId,
        IEnumerable<Permission> permissions)
    {
        var tasks = permissions.Select(p => 
            CheckPermissionAsync(userId, tenantId, p.Service, p.Entity, p.Action));
        
        var results = await Task.WhenAll(tasks);
        return results.Any(authorized => authorized);
    }
    
    public async Task<List<Permission>> GetUserPermissionsAsync(
        string userId,
        string tenantId)
    {
        var client = _httpClientFactory.CreateClient();
        var url = $"{_apiUrl}/api/v1/permissions/user?" +
                  $"user_id={userId}&tenant_id={tenantId}";
        
        var response = await client.GetFromJsonAsync<PermissionsResponse>(url);
        return response?.Data ?? new List<Permission>();
    }
}

public class Permission
{
    public string Service { get; set; }
    public string Entity { get; set; }
    public string Action { get; set; }
}

// Register in DI
builder.Services.AddScoped<IRbacService, RbacService>();

// Usage in controller
public class PostsController : ControllerBase
{
    private readonly IRbacService _rbac;
    
    public PostsController(IRbacService rbac)
    {
        _rbac = rbac;
    }
    
    [HttpPost]
    public async Task<IActionResult> CreatePost([FromBody] CreatePostRequest request)
    {
        var userId = User.FindFirst("sub")?.Value;
        var tenantId = RouteData.Values["tenantId"]?.ToString();
        
        var canCreate = await _rbac.CheckPermissionAsync(
            userId, tenantId, "blog-api", "post", "create");
        
        if (!canCreate)
        {
            return Forbid();
        }
        
        // Your logic
        return Ok(new { Success = true });
    }
}
```

## Planned API

When implemented, the API will follow this structure:

```csharp
// Configuration
services.AddUtmRbac(options =>
{
    options.ApiUrl = "http://localhost:8080";
    options.Timeout = TimeSpan.FromSeconds(30);
    options.CacheEnabled = true;
    options.CacheTTL = TimeSpan.FromMinutes(5);
});

// Middleware
app.UseUtmRbac();

// Attributes
[RequirePermission("blog-api", "post", "create")]
[RequireAnyPermission(permissions)]
[RequireTenantAccess]
[RequirePlatformAdmin]

// Direct checks
var rbac = serviceProvider.GetRequiredService<IRbacClient>();
var authorized = await rbac.CheckPermissionAsync(...);
var permissions = await rbac.GetUserPermissionsAsync(...);
```

## Package Structure

```
UtmRbac.AspNetCore/
├── Middleware/
│   ├── RbacMiddleware.cs
│   └── TenantAccessMiddleware.cs
├── Attributes/
│   ├── RequirePermissionAttribute.cs
│   ├── RequireAnyPermissionAttribute.cs
│   └── RequireTenantAccessAttribute.cs
├── Services/
│   ├── IRbacClient.cs
│   ├── RbacClient.cs
│   └── SystemUserClient.cs
├── Policies/
│   ├── PermissionRequirement.cs
│   └── PermissionHandler.cs
├── Extensions/
│   └── ServiceCollectionExtensions.cs
├── Models/
│   └── Permission.cs
└── Options/
    └── RbacOptions.cs
```

## Contributing

Interested in implementing this library? See:
- [Backend Integration Guide](/guides/backend-integration)
- [Go Middleware Implementation](/middleware/go)
- [RBAC API Reference](/x-api/rbac)

## Request for Implementation

If you need this library, please:
1. Star the repository
2. Open a GitHub issue with your use case
3. Consider contributing to implementation

**Implementation Priority**: Based on community demand

## Related Documentation

- [Go Middleware](/middleware/go) - Reference implementation
- [Python Middleware](/middleware/python) - Alternative language
- [Node.js Middleware](/middleware/nodejs) - Alternative language
- [System Auth Usage](/system-auth/usage) - System User examples
- [RBAC API](/x-api/rbac) - API reference
