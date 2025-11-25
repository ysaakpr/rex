# Java Middleware (Manual JWT Verification)

Since Java doesn't have an official SuperTokens SDK, we implement authentication by manually verifying JWT tokens.

## Overview

This implementation:
- ✅ Verifies JWT tokens from cookies or headers
- ✅ Supports both cookie-based and header-based auth
- ✅ Extracts user information from token claims
- ✅ Works with Spring Boot
- ✅ Detects system users from JWT payload

## Dependencies

Add to `pom.xml`:

```xml
<dependencies>
    <!-- JWT Library -->
    <dependency>
        <groupId>io.jsonwebtoken</groupId>
        <artifactId>jjwt-api</artifactId>
        <version>0.11.5</version>
    </dependency>
    <dependency>
        <groupId>io.jsonwebtoken</groupId>
        <artifactId>jjwt-impl</artifactId>
        <version>0.11.5</version>
        <scope>runtime</scope>
    </dependency>
    <dependency>
        <groupId>io.jsonwebtoken</groupId>
        <artifactId>jjwt-jackson</artifactId>
        <version>0.11.5</version>
        <scope>runtime</scope>
    </dependency>
    
    <!-- Servlet API -->
    <dependency>
        <groupId>javax.servlet</groupId>
        <artifactId>javax.servlet-api</artifactId>
        <version>4.0.1</version>
        <scope>provided</scope>
    </dependency>
</dependencies>
```

## Authentication Filter

Create `AuthFilter.java`:

```java
package com.example.middleware;

import io.jsonwebtoken.*;
import io.jsonwebtoken.security.Keys;
import javax.servlet.*;
import javax.servlet.http.*;
import java.io.IOException;
import java.security.Key;

/**
 * SuperTokens Authentication Filter for Java
 * Supports both cookie-based and header-based authentication
 */
public class AuthFilter implements Filter {
    
    private static final String ACCESS_TOKEN_COOKIE = "sAccessToken";
    private static final String AUTH_HEADER = "Authorization";
    private static final String ST_AUTH_MODE_HEADER = "st-auth-mode";
    
    private Key verificationKey;
    
    @Override
    public void init(FilterConfig filterConfig) throws ServletException {
        // Fetch JWT verification key
        String jwtSecret = System.getenv("SUPERTOKENS_JWT_SECRET");
        if (jwtSecret == null) {
            throw new ServletException("SUPERTOKENS_JWT_SECRET not configured");
        }
        this.verificationKey = Keys.hmacShaKeyFor(jwtSecret.getBytes());
    }
    
    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain chain)
            throws IOException, ServletException {
        
        HttpServletRequest httpRequest = (HttpServletRequest) request;
        HttpServletResponse httpResponse = (HttpServletResponse) response;
        
        try {
            // Determine auth mode
            String authMode = httpRequest.getHeader(ST_AUTH_MODE_HEADER);
            String token;
            
            if ("header".equals(authMode)) {
                // Header-based authentication
                token = extractTokenFromHeader(httpRequest);
            } else {
                // Cookie-based authentication (default)
                token = extractTokenFromCookie(httpRequest);
            }
            
            if (token == null) {
                sendUnauthorized(httpResponse, "No authentication token found");
                return;
            }
            
            // Verify and parse JWT
            Claims claims = verifyToken(token);
            
            // Extract user information
            String userId = claims.getSubject();
            httpRequest.setAttribute("userId", userId);
            
            // Check if system user
            Boolean isSystemUser = claims.get("is_system_user", Boolean.class);
            if (isSystemUser != null && isSystemUser) {
                httpRequest.setAttribute("isSystemUser", true);
                httpRequest.setAttribute("serviceName", 
                    claims.get("service_name", String.class));
                httpRequest.setAttribute("serviceType", 
                    claims.get("service_type", String.class));
            }
            
            // Store full claims for later use
            httpRequest.setAttribute("tokenClaims", claims);
            
            // Continue filter chain
            chain.doFilter(request, response);
            
        } catch (JwtException e) {
            sendUnauthorized(httpResponse, "Invalid token: " + e.getMessage());
        } catch (Exception e) {
            sendUnauthorized(httpResponse, "Authentication failed: " + e.getMessage());
        }
    }
    
    /**
     * Extract token from Authorization header
     */
    private String extractTokenFromHeader(HttpServletRequest request) {
        String authHeader = request.getHeader(AUTH_HEADER);
        if (authHeader != null && authHeader.startsWith("Bearer ")) {
            return authHeader.substring(7);
        }
        return null;
    }
    
    /**
     * Extract token from cookie
     */
    private String extractTokenFromCookie(HttpServletRequest request) {
        Cookie[] cookies = request.getCookies();
        if (cookies != null) {
            for (Cookie cookie : cookies) {
                if (ACCESS_TOKEN_COOKIE.equals(cookie.getName())) {
                    return cookie.getValue();
                }
            }
        }
        return null;
    }
    
    /**
     * Verify JWT token signature and expiry
     */
    private Claims verifyToken(String token) throws JwtException {
        return Jwts.parserBuilder()
                .setSigningKey(verificationKey)
                .build()
                .parseClaimsJws(token)
                .getBody();
    }
    
    /**
     * Send 401 Unauthorized response
     */
    private void sendUnauthorized(HttpServletResponse response, String message) 
            throws IOException {
        response.setStatus(HttpServletResponse.SC_UNAUTHORIZED);
        response.setContentType("application/json");
        response.getWriter().write(
            String.format("{\"success\": false, \"error\": \"%s\"}", message)
        );
    }
    
    @Override
    public void destroy() {
        // Cleanup if needed
    }
}
```

## Spring Boot Configuration

Register the filter in your Spring Boot application:

```java
package com.example.config;

import com.example.middleware.AuthFilter;
import org.springframework.boot.web.servlet.FilterRegistrationBean;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class SecurityConfig {
    
    @Bean
    public FilterRegistrationBean<AuthFilter> authFilter() {
        FilterRegistrationBean<AuthFilter> registrationBean = 
            new FilterRegistrationBean<>();
        
        registrationBean.setFilter(new AuthFilter());
        
        // Apply to all /api/v1/* endpoints
        registrationBean.addUrlPatterns("/api/v1/*");
        
        // Run early in filter chain
        registrationBean.setOrder(1);
        
        return registrationBean;
    }
}
```

## Using in Controllers

### Get Authenticated User

```java
@RestController
@RequestMapping("/api/v1")
public class TenantController {
    
    @GetMapping("/tenants")
    public ResponseEntity<?> listTenants(HttpServletRequest request) {
        // Get authenticated user ID
        String userId = (String) request.getAttribute("userId");
        
        // Check if system user
        Boolean isSystemUser = (Boolean) request.getAttribute("isSystemUser");
        if (isSystemUser != null && isSystemUser) {
            String serviceName = (String) request.getAttribute("serviceName");
            // Handle system user logic
        }
        
        // Your business logic
        List<Tenant> tenants = tenantService.listUserTenants(userId);
        return ResponseEntity.ok(tenants);
    }
}
```

### Helper Methods

Create a utility class for common operations:

```java
package com.example.util;

import javax.servlet.http.HttpServletRequest;
import java.util.Optional;

public class AuthUtils {
    
    /**
     * Get authenticated user ID from request
     */
    public static Optional<String> getUserId(HttpServletRequest request) {
        String userId = (String) request.getAttribute("userId");
        return Optional.ofNullable(userId);
    }
    
    /**
     * Check if request is from system user
     */
    public static boolean isSystemUser(HttpServletRequest request) {
        Boolean isSystemUser = (Boolean) request.getAttribute("isSystemUser");
        return isSystemUser != null && isSystemUser;
    }
    
    /**
     * Get service name (for system users)
     */
    public static Optional<String> getServiceName(HttpServletRequest request) {
        String serviceName = (String) request.getAttribute("serviceName");
        return Optional.ofNullable(serviceName);
    }
    
    /**
     * Get full JWT claims
     */
    public static io.jsonwebtoken.Claims getClaims(HttpServletRequest request) {
        return (io.jsonwebtoken.Claims) request.getAttribute("tokenClaims");
    }
}
```

Usage:

```java
@GetMapping("/protected")
public ResponseEntity<?> protectedEndpoint(HttpServletRequest request) {
    String userId = AuthUtils.getUserId(request)
        .orElseThrow(() -> new UnauthorizedException("Not authenticated"));
    
    if (AuthUtils.isSystemUser(request)) {
        String service = AuthUtils.getServiceName(request).orElse("unknown");
        // Handle system user
    }
    
    // Your logic
    return ResponseEntity.ok(data);
}
```

## Optional Authentication

For endpoints that work with or without authentication:

```java
public class OptionalAuthFilter implements Filter {
    
    // Same as AuthFilter, but don't return 401 on missing token
    
    @Override
    public void doFilter(ServletRequest request, ServletResponse response, 
                        FilterChain chain) throws IOException, ServletException {
        
        HttpServletRequest httpRequest = (HttpServletRequest) request;
        
        try {
            String token = extractToken(httpRequest);
            
            if (token != null) {
                Claims claims = verifyToken(token);
                httpRequest.setAttribute("userId", claims.getSubject());
            }
            // Continue even if no token
            chain.doFilter(request, response);
            
        } catch (Exception e) {
            // Log error but don't block request
            chain.doFilter(request, response);
        }
    }
}
```

## Fetching Public Key from SuperTokens

For better security, fetch the public key from SuperTokens JWKS endpoint:

```java
package com.example.security;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;

import java.math.BigInteger;
import java.security.KeyFactory;
import java.security.PublicKey;
import java.security.spec.RSAPublicKeySpec;
import java.util.Base64;

public class JWKSFetcher {
    
    private final String jwksUrl;
    private final OkHttpClient httpClient;
    private final ObjectMapper objectMapper;
    private PublicKey cachedKey;
    private long lastFetch = 0;
    private static final long CACHE_TTL = 3600000; // 1 hour
    
    public JWKSFetcher(String superTokensUrl) {
        this.jwksUrl = superTokensUrl + "/recipe/jwt/jwks";
        this.httpClient = new OkHttpClient();
        this.objectMapper = new ObjectMapper();
    }
    
    public PublicKey getPublicKey() throws Exception {
        // Return cached if still valid
        if (cachedKey != null && System.currentTimeMillis() - lastFetch < CACHE_TTL) {
            return cachedKey;
        }
        
        // Fetch from JWKS endpoint
        Request request = new Request.Builder()
                .url(jwksUrl)
                .build();
        
        try (Response response = httpClient.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new Exception("Failed to fetch JWKS: " + response.code());
            }
            
            String body = response.body().string();
            JsonNode jwks = objectMapper.readTree(body);
            JsonNode firstKey = jwks.get("keys").get(0);
            
            // Extract n and e from JWK
            String nStr = firstKey.get("n").asText();
            String eStr = firstKey.get("e").asText();
            
            // Decode Base64URL
            BigInteger modulus = new BigInteger(1, 
                Base64.getUrlDecoder().decode(nStr));
            BigInteger exponent = new BigInteger(1, 
                Base64.getUrlDecoder().decode(eStr));
            
            // Create public key
            RSAPublicKeySpec spec = new RSAPublicKeySpec(modulus, exponent);
            KeyFactory factory = KeyFactory.getInstance("RSA");
            cachedKey = factory.generatePublic(spec);
            lastFetch = System.currentTimeMillis();
            
            return cachedKey;
        }
    }
}
```

Update AuthFilter to use public key:

```java
private JWKSFetcher jwksFetcher;

@Override
public void init(FilterConfig filterConfig) throws ServletException {
    String superTokensUrl = System.getenv("SUPERTOKENS_CONNECTION_URI");
    this.jwksFetcher = new JWKSFetcher(superTokensUrl);
}

private Claims verifyToken(String token) throws Exception {
    PublicKey publicKey = jwksFetcher.getPublicKey();
    
    return Jwts.parserBuilder()
            .setSigningKey(publicKey)
            .build()
            .parseClaimsJws(token)
            .getBody();
}
```

## Testing

### Test with cURL

```bash
# Get access token (from frontend or login)
TOKEN="your-access-token"

# Test authentication
curl http://localhost:8080/api/v1/tenants \
  -H "Authorization: Bearer $TOKEN" \
  -H "st-auth-mode: header"
```

### Integration Test

```java
@SpringBootTest
@AutoConfigureMockMvc
public class AuthFilterTest {
    
    @Autowired
    private MockMvc mockMvc;
    
    @Test
    public void testUnauthorized() throws Exception {
        mockMvc.perform(get("/api/v1/tenants"))
                .andExpect(status().isUnauthorized());
    }
    
    @Test
    public void testWithValidToken() throws Exception {
        String token = createTestToken();
        
        mockMvc.perform(get("/api/v1/tenants")
                .header("Authorization", "Bearer " + token)
                .header("st-auth-mode", "header"))
                .andExpect(status().isOk());
    }
    
    private String createTestToken() {
        // Create test JWT token
        return Jwts.builder()
                .setSubject("test-user-id")
                .setExpiration(new Date(System.currentTimeMillis() + 3600000))
                .signWith(testKey)
                .compact();
    }
}
```

## Error Handling

Add custom exception handling:

```java
@ControllerAdvice
public class GlobalExceptionHandler {
    
    @ExceptionHandler(UnauthorizedException.class)
    public ResponseEntity<?> handleUnauthorized(UnauthorizedException ex) {
        return ResponseEntity
                .status(HttpStatus.UNAUTHORIZED)
                .body(Map.of("success", false, "error", ex.getMessage()));
    }
    
    @ExceptionHandler(ForbiddenException.class)
    public ResponseEntity<?> handleForbidden(ForbiddenException ex) {
        return ResponseEntity
                .status(HttpStatus.FORBIDDEN)
                .body(Map.of("success", false, "error", ex.getMessage()));
    }
}
```

## Complete Example

See the [full working example](https://github.com/yourusername/utm-backend/tree/main/examples/java-middleware) in the repository.

## Next Steps

- **[System Auth Library (Java)](/system-auth/java)** - Automated authentication for services
- **[C# Middleware](/middleware/csharp)** - Similar implementation for .NET
- **[API Reference](/x-api/overview)** - Explore available endpoints

