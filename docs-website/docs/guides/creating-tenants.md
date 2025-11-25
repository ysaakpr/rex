# Creating Tenants

Step-by-step guide to creating and initializing tenants.

## Two Creation Methods

### 1. Self-Service Onboarding

**Who**: Any authenticated user  
**When**: User creates their own workspace  
**Result**: User becomes tenant admin

### 2. Managed Onboarding

**Who**: Platform admins only  
**When**: Creating tenant for a customer  
**Result**: Customer receives invitation

## Self-Service Creation

### Step 1: Authenticate

User must be logged in:
```javascript
// Check authentication
import Session from "supertokens-auth-react/recipe/session";

const isAuthenticated = await Session.doesSessionExist();
```

### Step 2: Prepare Tenant Data

```javascript
const tenantData = {
  name: "My Company",           // Display name
  slug: "my-company",            // URL-friendly identifier
  metadata: {                    // Optional custom data
    industry: "Technology",
    size: "10-50",
    plan: "free"
  }
};
```

### Step 3: Create Tenant

```javascript
const response = await fetch('/api/v1/tenants', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify(tenantData)
});

const result = await response.json();
console.log(result.data.id); // Tenant UUID
```

### Step 4: Wait for Initialization

```javascript
// Tenant status starts as "pending"
const checkStatus = async (tenantId) => {
  const response = await fetch(`/api/v1/tenants/${tenantId}/status`, {
    credentials: 'include'
  });
  const data = await response.json();
  return data.data.status; // "pending" or "active"
};

// Poll until active
while (await checkStatus(tenantId) === 'pending') {
  await new Promise(resolve => setTimeout(resolve, 2000));
}
```

### What Happens Automatically

1. ✅ Tenant record created (status: pending)
2. ✅ Creator added as Admin member
3. ✅ Background job enqueued
4. ✅ Initialization job runs:
   - Set up default resources
   - Configure integrations
   - Send welcome email
5. ✅ Status changed to "active"

## Managed Creation (Platform Admin)

### Step 1: Authenticate as Admin

Must be platform admin:
```bash
GET /api/v1/platform/admins/check
# Should return: {"is_admin": true}
```

### Step 2: Create Managed Tenant

```javascript
const tenantData = {
  name: "Enterprise Customer Inc",
  slug: "enterprise-customer",
  admin_email: "admin@enterprise-customer.com",  // Customer's email
  metadata: {
    plan: "enterprise",
    contract_value: 50000,
    sales_rep: "John Doe"
  }
};

const response = await fetch('/api/v1/tenants/managed', {
  method: 'POST',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify(tenantData)
});
```

### Step 3: Invitation Sent

System automatically:
1. Creates tenant
2. Generates invitation
3. Sends email to `admin_email`
4. Customer clicks link
5. Customer signs up/logs in
6. Customer accepts invitation
7. Customer becomes tenant admin

## Slug Requirements

### Valid Slugs

```
✅ "acme-corp"
✅ "my-startup"
✅ "company-123"
✅ "tech-co"
```

### Invalid Slugs

```
❌ "Acme Corp"        # Spaces not allowed
❌ "acme_corp"        # Underscores not allowed
❌ "acme.corp"        # Dots not allowed
❌ "a"                # Too short (min 3 chars)
```

### Slug Validation

```go
// Backend validation
slug: "required,min=3,max=255,alphanum"

// Alphanumeric + hyphens only
// Must be unique across all tenants
```

### Best Practices

1. **Lowercase**: Always use lowercase
2. **Short**: Keep under 30 characters
3. **Descriptive**: Reflect company name
4. **Permanent**: Don't change after creation

## Metadata Usage

### Common Metadata Fields

```json
{
  "metadata": {
    // Business info
    "industry": "Technology",
    "company_size": "50-100",
    
    // Plan info
    "plan": "enterprise",
    "billing_cycle": "annual",
    
    // Custom fields
    "crm_id": "SF-12345",
    "sales_rep": "John Doe",
    
    // Features
    "features": {
      "advanced_analytics": true,
      "custom_branding": true,
      "api_access": true
    }
  }
}
```

### Accessing Metadata

```javascript
// Frontend
const tenant = await getTenant(tenantId);
const plan = tenant.metadata.plan;
const features = tenant.metadata.features;

if (features.advanced_analytics) {
  // Show analytics dashboard
}
```

### Updating Metadata

```javascript
await fetch(`/api/v1/tenants/${tenantId}`, {
  method: 'PATCH',
  credentials: 'include',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    metadata: {
      ...existingMetadata,
      plan: "pro",  // Upgrade plan
      upgraded_at: new Date().toISOString()
    }
  })
});
```

## Complete Example (React)

```javascript
import { useState } from 'react';

function CreateTenantForm() {
  const [loading, setLoading] = useState(false);
  const [tenant, setTenant] = useState(null);
  
  const createTenant = async (e) => {
    e.preventDefault();
    setLoading(true);
    
    const formData = new FormData(e.target);
    const data = {
      name: formData.get('name'),
      slug: formData.get('slug').toLowerCase(),
      metadata: {
        industry: formData.get('industry')
      }
    };
    
    try {
      const response = await fetch('/api/v1/tenants', {
        method: 'POST',
        credentials: 'include',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(data)
      });
      
      if (!response.ok) {
        const error = await response.json();
        alert(error.error);
        return;
      }
      
      const result = await response.json();
      setTenant(result.data);
      
      // Wait for initialization
      await waitForActive(result.data.id);
      
      // Redirect to tenant dashboard
      window.location.href = `/tenants/${result.data.id}`;
      
    } catch (error) {
      alert('Failed to create tenant');
    } finally {
      setLoading(false);
    }
  };
  
  const waitForActive = async (tenantId) => {
    let attempts = 0;
    while (attempts < 30) {
      const status = await checkStatus(tenantId);
      if (status === 'active') return;
      await new Promise(r => setTimeout(r, 1000));
      attempts++;
    }
  };
  
  return (
    <form onSubmit={createTenant}>
      <input name="name" placeholder="Company Name" required />
      <input name="slug" placeholder="company-slug" required />
      <input name="industry" placeholder="Industry" />
      <button disabled={loading}>
        {loading ? 'Creating...' : 'Create Tenant'}
      </button>
    </form>
  );
}
```

## Next Steps

- [Member Management](/guides/member-management) - Add team members
- [Invitations](/guides/invitations) - Invite users to join
- [Tenants API](/x-api/tenants) - API reference
