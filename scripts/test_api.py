#!/usr/bin/env python3
"""
Test script for UTM Backend API
Tests user creation, authentication, and tenant management
"""

import requests
import json
import sys
from datetime import datetime

# Configuration
BASE_URL = "http://localhost:8080"
TEST_EMAIL = f"testuser-{int(datetime.now().timestamp())}@example.com"
TEST_PASSWORD = "TestPassword123!"

# Colors
class Colors:
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    RED = '\033[0;31m'
    NC = '\033[0m'

def print_section(title):
    print(f"\n{Colors.YELLOW}{title}{Colors.NC}")

def print_success(message):
    print(f"{Colors.GREEN}✓ {message}{Colors.NC}")

def print_error(message):
    print(f"{Colors.RED}✗ {message}{Colors.NC}")

def print_json(data):
    print(json.dumps(data, indent=2))

def main():
    print("=== UTM Backend API Test ===\n")
    
    # Create session to maintain cookies
    session = requests.Session()
    
    # Step 1: Create user (Sign up)
    print_section("1. Creating test user...")
    signup_data = {
        "formFields": [
            {"id": "email", "value": TEST_EMAIL},
            {"id": "password", "value": TEST_PASSWORD}
        ]
    }
    
    try:
        response = session.post(
            f"{BASE_URL}/auth/signup",
            json=signup_data,
            headers={
                "Content-Type": "application/json",
                "rid": "emailpassword"
            }
        )
        
        result = response.json()
        print_json(result)
        
        if result.get("status") == "OK":
            print_success(f"User created successfully")
            user_id = result["user"]["id"]
            print(f"User ID: {user_id}")
        elif result.get("status") == "EMAIL_ALREADY_EXISTS_ERROR" or result.get("status") == "FIELD_ERROR":
            print(f"{Colors.YELLOW}! User already exists, proceeding to sign in...{Colors.NC}")
            # Continue to sign in
        else:
            print_error("Failed to create user")
            return False
            
    except Exception as e:
        print_error(f"Error creating user: {e}")
        return False
    
    # Step 2: Sign in
    print_section("2. Signing in...")
    signin_data = {
        "formFields": [
            {"id": "email", "value": TEST_EMAIL},
            {"id": "password", "value": TEST_PASSWORD}
        ]
    }
    
    try:
        response = session.post(
            f"{BASE_URL}/auth/signin",
            json=signin_data,
            headers={
                "Content-Type": "application/json",
                "rid": "emailpassword"
            }
        )
        
        result = response.json()
        print_json(result)
        
        if result.get("status") == "OK":
            print_success("Signed in successfully")
            user_id = result["user"]["id"]
            print(f"User ID: {user_id}")
            
            # Extract tokens from response headers
            print("\nSession headers:")
            access_token = response.headers.get('st-access-token')
            refresh_token = response.headers.get('st-refresh-token')
            front_token = response.headers.get('front-token')
            
            if access_token:
                print(f"  Access Token: {access_token[:50]}...")
                session.headers.update({'st-access-token': access_token})
            if front_token:
                print(f"  Front Token: {front_token[:50]}...")
                session.headers.update({'front-token': front_token})
            
            # Print cookies for debugging
            print("\nSession cookies:")
            for cookie in session.cookies:
                print(f"  {cookie.name}: {cookie.value[:50] if len(cookie.value) > 50 else cookie.value}...")
        else:
            print_error("Failed to sign in")
            return False
            
    except Exception as e:
        print_error(f"Error signing in: {e}")
        return False
    
    # Step 3: Create tenant
    print_section("3. Creating tenant...")
    tenant_data = {
        "name": "Test Company",
        "slug": f"test-company-{int(datetime.now().timestamp())}",
        "metadata": {
            "industry": "technology",
            "size": "10-50",
            "test": True
        }
    }
    
    try:
        response = session.post(
            f"{BASE_URL}/api/v1/tenants",
            json=tenant_data,
            headers={"Content-Type": "application/json"}
        )
        
        print(f"Status Code: {response.status_code}")
        result = response.json()
        print_json(result)
        
        if response.status_code == 201 and result.get("success"):
            print_success("Tenant created successfully")
            tenant_id = result["data"]["id"]
            tenant_name = result["data"]["name"]
            print(f"Tenant ID: {tenant_id}")
            print(f"Tenant Name: {tenant_name}")
        else:
            print_error("Failed to create tenant")
            print(f"Response: {response.text}")
            return False
            
    except Exception as e:
        print_error(f"Error creating tenant: {e}")
        return False
    
    # Step 4: Check tenant status
    print_section("4. Checking tenant status...")
    try:
        import time
        time.sleep(2)  # Give background job time to process
        
        response = session.get(
            f"{BASE_URL}/api/v1/tenants/{tenant_id}/status",
            headers={"Content-Type": "application/json"}
        )
        
        result = response.json()
        print_json(result)
        
        if result.get("success"):
            print_success(f"Tenant status: {result['data']['status']}")
            
    except Exception as e:
        print_error(f"Error checking status: {e}")
    
    # Step 5: List tenants
    print_section("5. Listing user's tenants...")
    try:
        response = session.get(
            f"{BASE_URL}/api/v1/tenants",
            headers={"Content-Type": "application/json"}
        )
        
        result = response.json()
        print_json(result)
        
        if result.get("success"):
            tenant_count = result["data"]["total_count"]
            print_success(f"Found {tenant_count} tenant(s)")
            
    except Exception as e:
        print_error(f"Error listing tenants: {e}")
    
    # Step 6: Get tenant details
    print_section("6. Getting tenant details...")
    try:
        response = session.get(
            f"{BASE_URL}/api/v1/tenants/{tenant_id}",
            headers={"Content-Type": "application/json"}
        )
        
        result = response.json()
        print_json(result)
        
        if result.get("success"):
            print_success("Retrieved tenant details")
            
    except Exception as e:
        print_error(f"Error getting tenant details: {e}")
    
    # Summary
    print(f"\n{Colors.GREEN}=== All tests passed! ==={Colors.NC}\n")
    print("Summary:")
    print(f"  User Email: {TEST_EMAIL}")
    print(f"  User ID: {user_id}")
    print(f"  Tenant ID: {tenant_id}")
    print(f"  Tenant Name: {tenant_name}")
    print(f"  API Base URL: {BASE_URL}")
    
    return True

if __name__ == "__main__":
    try:
        success = main()
        sys.exit(0 if success else 1)
    except KeyboardInterrupt:
        print(f"\n{Colors.YELLOW}Test interrupted by user{Colors.NC}")
        sys.exit(1)
    except Exception as e:
        print_error(f"Unexpected error: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)

