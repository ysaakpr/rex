# Java System Auth Library

Complete System User Authentication Library for Java applications.

## Overview

Since Java doesn't have an official SuperTokens SDK, this library provides:
- ✅ Automatic login and token management
- ✅ Token refresh before expiry  
- ✅ 401 error recovery with re-login
- ✅ Pluggable secret storage (vaults)
- ✅ Thread-safe operations

## Installation

Add to your `pom.xml`:

```xml
<dependencies>
    <!-- HTTP Client -->
    <dependency>
        <groupId>com.squareup.okhttp3</groupId>
        <artifactId>okhttp</artifactId>
        <version>4.11.0</version>
    </dependency>
    
    <!-- JSON -->
    <dependency>
        <groupId>com.google.code.gson</groupId>
        <artifactId>gson</artifactId>
        <version>2.10.1</version>
    </dependency>
</dependencies>
```

## Quick Start

```java
import com.yourorg.systemauth.*;

// 1. Create vault
SecretVault vault = new EnvVault("WORKER_EMAIL", "WORKER_PASSWORD");

// 2. Configure client
SystemAuthClient.Config config = new SystemAuthClient.Config();
config.vault = vault;
config.apiUrl = "https://api.yourdomain.com";

// 3. Create client
SystemAuthClient client = new SystemAuthClient(config);

// 4. Make authenticated requests (login automatic!)
Response response = client.makeAuthenticatedRequest(
    "GET",
    "https://api.yourdomain.com/api/v1/tenants",
    null
);

System.out.println(response.body().string());
```

## Vault Interface

```java
public interface SecretVault {
    String getEmail() throws Exception;
    String getPassword() throws Exception;
}
```

### EnvVault Implementation

```java
public class EnvVault implements SecretVault {
    private final String emailEnvKey;
    private final String passwordEnvKey;
    
    public EnvVault(String emailEnvKey, String passwordEnvKey) {
        this.emailEnvKey = emailEnvKey;
        this.passwordEnvKey = passwordEnvKey;
    }
    
    @Override
    public String getEmail() throws Exception {
        String email = System.getenv(emailEnvKey);
        if (email == null || email.isEmpty()) {
            throw new Exception("Email not found in environment: " + emailEnvKey);
        }
        return email;
    }
    
    @Override
    public String getPassword() throws Exception {
        String password = System.getenv(passwordEnvKey);
        if (password == null || password.isEmpty()) {
            throw new Exception("Password not found in environment: " + passwordEnvKey);
        }
        return password;
    }
}
```

**Usage**:
```java
SecretVault vault = new EnvVault("SYSTEM_USER_EMAIL", "SYSTEM_USER_PASSWORD");
```

**Environment Variables**:
```bash
export SYSTEM_USER_EMAIL="worker@system.internal"
export SYSTEM_USER_PASSWORD="sysuser_abc123..."
```

## SystemAuthClient

### Configuration

```java
public static class Config {
    public SecretVault vault;           // Required
    public String apiUrl;                // Required
    public int refreshBufferSeconds = 300; // 5 minutes before expiry
    public int connectTimeoutSeconds = 30;
    public int readTimeoutSeconds = 30;
}
```

### Full Implementation

```java
import okhttp3.*;
import com.google.gson.*;
import java.util.concurrent.locks.ReentrantLock;

public class SystemAuthClient {
    private final Config config;
    private final OkHttpClient httpClient;
    private final Gson gson;
    private final ReentrantLock lock;
    
    // Token storage
    private String accessToken;
    private long tokenExpiryTime;
    
    public static class Config {
        public SecretVault vault;
        public String apiUrl;
        public int refreshBufferSeconds = 300;
        public int connectTimeoutSeconds = 30;
        public int readTimeoutSeconds = 30;
    }
    
    public SystemAuthClient(Config config) {
        this.config = config;
        this.gson = new Gson();
        this.lock = new ReentrantLock();
        
        this.httpClient = new OkHttpClient.Builder()
            .connectTimeout(config.connectTimeoutSeconds, TimeUnit.SECONDS)
            .readTimeout(config.readTimeoutSeconds, TimeUnit.SECONDS)
            .build();
    }
    
    /**
     * Make authenticated API request
     * Handles login and token refresh automatically
     */
    public Response makeAuthenticatedRequest(
            String method, 
            String url, 
            String jsonBody) throws Exception {
        
        // Ensure we have valid token
        ensureValidToken();
        
        // Build request
        Request.Builder requestBuilder = new Request.Builder()
            .url(url)
            .addHeader("Authorization", "Bearer " + accessToken)
            .addHeader("st-auth-mode", "header");
        
        if (jsonBody != null) {
            RequestBody body = RequestBody.create(
                jsonBody,
                MediaType.parse("application/json")
            );
            requestBuilder.method(method, body);
        } else {
            requestBuilder.method(method, null);
        }
        
        // Execute request
        Response response = httpClient.newCall(requestBuilder.build()).execute();
        
        // Handle 401 - token might be invalid
        if (response.code() == 401) {
            response.close();
            
            // Re-authenticate and retry
            lock.lock();
            try {
                accessToken = null; // Force re-login
                ensureValidToken();
            } finally {
                lock.unlock();
            }
            
            // Retry request with new token
            requestBuilder.header("Authorization", "Bearer " + accessToken);
            response = httpClient.newCall(requestBuilder.build()).execute();
        }
        
        return response;
    }
    
    /**
     * Ensure we have valid access token
     * Thread-safe
     */
    private void ensureValidToken() throws Exception {
        lock.lock();
        try {
            long now = System.currentTimeMillis() / 1000;
            long bufferTime = config.refreshBufferSeconds;
            
            if (accessToken == null || (now + bufferTime) >= tokenExpiryTime) {
                login();
            }
        } finally {
            lock.unlock();
        }
    }
    
    /**
     * Login to get access token
     */
    private void login() throws Exception {
        String email = config.vault.getEmail();
        String password = config.vault.getPassword();
        
        // Build login request
        JsonObject loginData = new JsonObject();
        JsonArray formFields = new JsonArray();
        
        JsonObject emailField = new JsonObject();
        emailField.addProperty("id", "email");
        emailField.addProperty("value", email);
        formFields.add(emailField);
        
        JsonObject passwordField = new JsonObject();
        passwordField.addProperty("id", "password");
        passwordField.addProperty("value", password);
        formFields.add(passwordField);
        
        loginData.add("formFields", formFields);
        
        // Send login request
        RequestBody body = RequestBody.create(
            gson.toJson(loginData),
            MediaType.parse("application/json")
        );
        
        Request request = new Request.Builder()
            .url(config.apiUrl + "/api/auth/signin")
            .post(body)
            .addHeader("Content-Type", "application/json")
            .build();
        
        try (Response response = httpClient.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new Exception("Login failed with status: " + response.code());
            }
            
            // Extract access token from header
            String tokenHeader = response.header("st-access-token");
            if (tokenHeader == null || tokenHeader.isEmpty()) {
                throw new Exception("No access token in response");
            }
            
            accessToken = tokenHeader;
            
            // Calculate expiry (24 hours for system users)
            tokenExpiryTime = (System.currentTimeMillis() / 1000) + (24 * 3600);
        }
    }
    
    /**
     * Explicitly login (optional, automatic otherwise)
     */
    public void explicitLogin() throws Exception {
        lock.lock();
        try {
            login();
        } finally {
            lock.unlock();
        }
    }
    
    /**
     * Get current access token
     */
    public String getAccessToken() throws Exception {
        ensureValidToken();
        return accessToken;
    }
}
```

## Custom Vault: AWS Secrets Manager

```java
import com.amazonaws.services.secretsmanager.*;
import com.amazonaws.services.secretsmanager.model.*;
import com.google.gson.*;

public class AWSSecretsVault implements SecretVault {
    private final String secretName;
    private final String region;
    private JsonObject cachedSecret;
    
    public AWSSecretsVault(String secretName, String region) {
        this.secretName = secretName;
        this.region = region;
    }
    
    private JsonObject getSecret() throws Exception {
        if (cachedSecret != null) {
            return cachedSecret;
        }
        
        AWSSecretsManager client = AWSSecretsManagerClientBuilder
            .standard()
            .withRegion(region)
            .build();
        
        GetSecretValueRequest request = new GetSecretValueRequest()
            .withSecretId(secretName);
        
        GetSecretValueResult result = client.getSecretValue(request);
        String secretString = result.getSecretString();
        
        cachedSecret = new Gson().fromJson(secretString, JsonObject.class);
        return cachedSecret;
    }
    
    @Override
    public String getEmail() throws Exception {
        return getSecret().get("email").getAsString();
    }
    
    @Override
    public String getPassword() throws Exception {
        return getSecret().get("password").getAsString();
    }
}
```

**Usage**:
```java
SecretVault vault = new AWSSecretsVault("utm-worker-credentials", "us-east-1");
```

## Complete Example: Background Worker

```java
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

public class BackgroundWorker {
    private final SystemAuthClient authClient;
    private final ScheduledExecutorService scheduler;
    
    public BackgroundWorker() {
        // Setup auth client
        SecretVault vault = new EnvVault("WORKER_EMAIL", "WORKER_PASSWORD");
        
        SystemAuthClient.Config config = new SystemAuthClient.Config();
        config.vault = vault;
        config.apiUrl = System.getenv("API_URL");
        
        this.authClient = new SystemAuthClient(config);
        this.scheduler = Executors.newScheduledThreadPool(1);
    }
    
    public void start() {
        System.out.println("Worker started");
        
        // Run every 10 seconds
        scheduler.scheduleAtFixedRate(
            this::processJobs,
            0,
            10,
            TimeUnit.SECONDS
        );
    }
    
    private void processJobs() {
        try {
            // Fetch pending jobs (authentication automatic)
            Response response = authClient.makeAuthenticatedRequest(
                "GET",
                System.getenv("API_URL") + "/api/v1/jobs/pending",
                null
            );
            
            if (!response.isSuccessful()) {
                System.err.println("Failed to fetch jobs: " + response.code());
                return;
            }
            
            String body = response.body().string();
            // Parse and process jobs...
            
        } catch (Exception e) {
            System.err.println("Error processing jobs: " + e.getMessage());
        }
    }
    
    public static void main(String[] args) {
        BackgroundWorker worker = new BackgroundWorker();
        worker.start();
        
        // Keep running
        try {
            Thread.currentThread().join();
        } catch (InterruptedException e) {
            System.exit(0);
        }
    }
}
```

## Example: Scheduled Task

```java
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

@Service
public class DailyReportService {
    private final SystemAuthClient authClient;
    
    public DailyReportService() {
        SecretVault vault = new EnvVault("CRON_EMAIL", "CRON_PASSWORD");
        
        SystemAuthClient.Config config = new SystemAuthClient.Config();
        config.vault = vault;
        config.apiUrl = System.getenv("API_URL");
        
        this.authClient = new SystemAuthClient(config);
    }
    
    @Scheduled(cron = "0 2 * * *") // Daily at 2 AM
    public void generateDailyReport() {
        try {
            System.out.println("Generating daily report");
            
            JsonObject reportData = new JsonObject();
            reportData.addProperty("report_type", "daily_summary");
            reportData.addProperty("date", LocalDate.now().toString());
            
            Response response = authClient.makeAuthenticatedRequest(
                "POST",
                System.getenv("API_URL") + "/api/v1/reports/generate",
                new Gson().toJson(reportData)
            );
            
            if (response.isSuccessful()) {
                System.out.println("Report generated successfully");
            } else {
                System.err.println("Report generation failed: " + response.code());
            }
            
        } catch (Exception e) {
            System.err.println("Error generating report: " + e.getMessage());
        }
    }
}
```

## Testing

### Unit Test with Mock Vault

```java
import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

public class SystemAuthClientTest {
    
    static class MockVault implements SecretVault {
        @Override
        public String getEmail() {
            return "test@system.internal";
        }
        
        @Override
        public String getPassword() {
            return "test_password";
        }
    }
    
    @Test
    public void testAuthentication() {
        SystemAuthClient.Config config = new SystemAuthClient.Config();
        config.vault = new MockVault();
        config.apiUrl = "http://test-api";
        
        SystemAuthClient client = new SystemAuthClient(config);
        
        // Test that client initializes
        assertNotNull(client);
    }
}
```

## Deployment

### Docker

**Dockerfile**:
```dockerfile
FROM maven:3.8-openjdk-17 AS builder
WORKDIR /app
COPY pom.xml .
COPY src ./src
RUN mvn clean package

FROM openjdk:17-slim
WORKDIR /app
COPY --from=builder /app/target/worker.jar .
CMD ["java", "-jar", "worker.jar"]
```

**docker-compose.yml**:
```yaml
services:
  worker:
    build: .
    environment:
      - WORKER_EMAIL=worker@system.internal
      - WORKER_PASSWORD=${WORKER_PASSWORD}
      - API_URL=https://api.yourdomain.com
    restart: unless-stopped
```

## Next Steps

- [Go Implementation](/system-auth/go) - Go version
- [Custom Vaults](/system-auth/custom-vaults) - More vault implementations
- [System Users Guide](/guides/system-users) - M2M authentication
