import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { Mail, Building2, UserPlus, Loader2, CheckCircle, AlertTriangle } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import Session from 'supertokens-auth-react/recipe/session';

export function AcceptInvitationPage() {
  const { token: tokenParam } = useParams();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  
  // Token can be either from URL params or query string
  const token = tokenParam || searchParams.get('token');
  
  const [loading, setLoading] = useState(true);
  const [accepting, setAccepting] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [invitation, setInvitation] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [currentUserEmail, setCurrentUserEmail] = useState(null);
  const [emailMismatch, setEmailMismatch] = useState(false);

  useEffect(() => {
    if (token) {
      checkAuthAndLoadInvitation();
    } else {
      setError('Invalid invitation link - no token provided');
      setLoading(false);
    }
  }, [token]);

  const checkAuthAndLoadInvitation = async () => {
    try {
      setLoading(true);
      setError('');
      setEmailMismatch(false);

      // Check if user is authenticated
      const sessionExists = await Session.doesSessionExist();
      setIsAuthenticated(sessionExists);

      // Load invitation details
      const invitationResponse = await fetch(`/api/v1/invitations/${token}`, {
        credentials: 'include'
      });

      if (!invitationResponse.ok) {
        const errorData = await invitationResponse.json().catch(() => ({}));
        const errorMessage = errorData.error || errorData.message;
        
        if (invitationResponse.status === 404) {
          throw new Error(errorMessage || 'Invitation not found or has expired');
        }
        if (invitationResponse.status === 400) {
          // Handle specific error messages from backend
          throw new Error(errorMessage || 'Invalid invitation');
        }
        throw new Error(errorMessage || 'Failed to load invitation details');
      }

      const invitationData = await invitationResponse.json();
      setInvitation(invitationData.data);

      // If authenticated, check if email matches
      if (sessionExists) {
        try {
          const userResponse = await fetch('/api/v1/users/me', {
            credentials: 'include'
          });
          
          if (userResponse.ok) {
            const userData = await userResponse.json();
            const userEmail = userData.data?.email;
            setCurrentUserEmail(userEmail);

            // Check if emails match
            if (userEmail && invitationData.data.email) {
              const emailsMatch = userEmail.toLowerCase() === invitationData.data.email.toLowerCase();
              setEmailMismatch(!emailsMatch);
              
              if (!emailsMatch) {
                setError(`This invitation was sent to ${invitationData.data.email}, but you are logged in as ${userEmail}. Please log in with the correct account.`);
              }
            }
          }
        } catch (err) {
          console.error('Error fetching user email:', err);
          // Don't fail the whole flow if we can't get the email
        }
      }
    } catch (err) {
      console.error('Error loading invitation:', err);
      setError(err.message || 'Failed to load invitation');
    } finally {
      setLoading(false);
    }
  };

  const handleAcceptInvitation = async () => {
    if (!isAuthenticated) {
      // Redirect to sign in with return URL (including query params)
      const returnUrl = encodeURIComponent(window.location.pathname + window.location.search);
      navigate(`/auth/signin?redirectToPath=${returnUrl}`);
      return;
    }

    try {
      setAccepting(true);
      setError('');

      const response = await fetch(`/api/v1/invitations/${token}/accept`, {
        method: 'POST',
        credentials: 'include'
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to accept invitation');
      }

      setSuccess(true);

      // Redirect to tenant after a short delay
      setTimeout(() => {
        navigate('/tenants');
      }, 2000);
    } catch (err) {
      console.error('Error accepting invitation:', err);
      setError(err.message || 'Failed to accept invitation');
    } finally {
      setAccepting(false);
    }
  };

  const handleSignUp = () => {
    const returnUrl = encodeURIComponent(window.location.pathname + window.location.search);
    navigate(`/auth?redirectToPath=${returnUrl}`);
  };

  const handleSignIn = () => {
    const returnUrl = encodeURIComponent(window.location.pathname + window.location.search);
    navigate(`/auth/signin?redirectToPath=${returnUrl}`);
  };

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-muted/50">
        <div className="text-center">
          <Loader2 className="h-12 w-12 animate-spin text-primary mx-auto" />
          <p className="mt-4 text-muted-foreground">Loading invitation...</p>
        </div>
      </div>
    );
  }

  if (error && !invitation) {
    const isAlreadyAccepted = error.toLowerCase().includes('already been accepted');
    const isExpired = error.toLowerCase().includes('expired');
    const isCancelled = error.toLowerCase().includes('cancelled');
    
    return (
      <div className="flex h-screen items-center justify-center bg-muted/50">
        <Card className="w-full max-w-md">
          <CardHeader>
            <div className="flex items-center justify-center mb-4">
              <div className={`flex h-16 w-16 items-center justify-center rounded-full ${
                isAlreadyAccepted ? 'bg-blue-100 dark:bg-blue-900/30' : 'bg-destructive/10'
              }`}>
                {isAlreadyAccepted ? (
                  <CheckCircle className="h-8 w-8 text-blue-600 dark:text-blue-400" />
                ) : (
                  <AlertTriangle className="h-8 w-8 text-destructive" />
                )}
              </div>
            </div>
            <CardTitle className="text-center">
              {isAlreadyAccepted ? 'Invitation Already Accepted' : 
               isExpired ? 'Invitation Expired' :
               isCancelled ? 'Invitation Cancelled' :
               'Invitation Invalid'}
            </CardTitle>
            <CardDescription className="text-center">
              {isAlreadyAccepted ? (
                <>
                  This invitation has already been accepted. If you were the one who accepted it, you should already have access to the tenant.
                </>
              ) : (
                error
              )}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {isAlreadyAccepted && (
              <Button 
                onClick={() => navigate('/tenants')} 
                className="w-full"
              >
                Go to My Tenants
              </Button>
            )}
            <Button 
              onClick={() => navigate('/')} 
              className="w-full"
              variant="outline"
            >
              Go to Home
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (success) {
    return (
      <div className="flex h-screen items-center justify-center bg-muted/50">
        <Card className="w-full max-w-md">
          <CardHeader>
            <div className="flex items-center justify-center mb-4">
              <div className="flex h-16 w-16 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30">
                <CheckCircle className="h-8 w-8 text-green-600 dark:text-green-400" />
              </div>
            </div>
            <CardTitle className="text-center">Welcome Aboard!</CardTitle>
            <CardDescription className="text-center">
              You've successfully joined {invitation?.tenant?.name}
            </CardDescription>
          </CardHeader>
          <CardContent className="text-center">
            <p className="text-sm text-muted-foreground mb-4">
              Redirecting you to your dashboard...
            </p>
            <Loader2 className="h-6 w-6 animate-spin text-primary mx-auto" />
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/50 p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <div className="flex items-center justify-center mb-4">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
              <Mail className="h-8 w-8 text-primary" />
            </div>
          </div>
          <CardTitle className="text-center">You're Invited!</CardTitle>
          <CardDescription className="text-center">
            You've been invited to join a tenant
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Invitation Details */}
          <div className="space-y-4">
            <div className="flex items-start gap-3 p-4 border rounded-lg bg-muted/50">
              <Building2 className="h-5 w-5 text-primary mt-0.5" />
              <div className="flex-1">
                <p className="font-medium">{invitation?.tenant?.name}</p>
                <p className="text-sm text-muted-foreground">{invitation?.tenant?.slug}</p>
              </div>
            </div>

            <div className="flex items-start gap-3 p-4 border rounded-lg bg-muted/50">
              <UserPlus className="h-5 w-5 text-primary mt-0.5" />
              <div className="flex-1">
                <p className="text-sm font-medium">Role</p>
                <Badge variant="outline" className="mt-1">
                  {invitation?.role?.name || 'Member'}
                </Badge>
              </div>
            </div>

            <div className="flex items-start gap-3 p-4 border rounded-lg bg-muted/50">
              <Mail className="h-5 w-5 text-primary mt-0.5" />
              <div className="flex-1">
                <p className="text-sm font-medium">Invited Email</p>
                <p className="text-sm text-muted-foreground mt-1">{invitation?.email}</p>
              </div>
            </div>
          </div>

          {/* Error Message */}
          {error && (
            <div className="rounded-lg bg-destructive/10 border border-destructive/20 p-3">
              <p className="text-sm text-destructive">{error}</p>
            </div>
          )}

          {/* Action Buttons */}
          {isAuthenticated ? (
            <div className="space-y-3">
              {emailMismatch ? (
                <>
                  <div className="rounded-lg bg-yellow-50 dark:bg-yellow-950 border border-yellow-200 dark:border-yellow-800 p-4">
                    <div className="flex gap-3">
                      <AlertTriangle className="h-5 w-5 text-yellow-600 dark:text-yellow-400 flex-shrink-0 mt-0.5" />
                      <div className="text-sm text-yellow-800 dark:text-yellow-200">
                        <p className="font-medium mb-1">Wrong Account</p>
                        <p>You're logged in as <strong>{currentUserEmail}</strong>, but this invitation was sent to <strong>{invitation?.email}</strong>.</p>
                      </div>
                    </div>
                  </div>
                  <Button 
                    onClick={() => {
                      // Log out and redirect to sign in
                      Session.signOut();
                      navigate('/auth/signin');
                    }}
                    variant="outline"
                    className="w-full gap-2"
                  >
                    Sign Out & Login with {invitation?.email}
                  </Button>
                </>
              ) : (
                <>
                  <Button 
                    onClick={handleAcceptInvitation} 
                    className="w-full gap-2"
                    disabled={accepting}
                  >
                    {accepting ? (
                      <>
                        <Loader2 className="h-4 w-4 animate-spin" />
                        Accepting...
                      </>
                    ) : (
                      <>
                        <CheckCircle className="h-4 w-4" />
                        Accept Invitation
                      </>
                    )}
                  </Button>
                  <Button 
                    onClick={() => navigate('/')} 
                    variant="outline"
                    className="w-full"
                    disabled={accepting}
                  >
                    Decline
                  </Button>
                </>
              )}
            </div>
          ) : (
            <div className="space-y-3">
              <div className="text-center text-sm text-muted-foreground mb-4">
                To accept this invitation, please sign in or create an account
              </div>
              <Button 
                onClick={handleSignUp} 
                className="w-full gap-2"
              >
                <UserPlus className="h-4 w-4" />
                Create Account
              </Button>
              <Button 
                onClick={handleSignIn} 
                variant="outline"
                className="w-full"
              >
                Already have an account? Sign In
              </Button>
            </div>
          )}

          <div className="text-center text-xs text-muted-foreground">
            This invitation will expire on {new Date(invitation?.expires_at).toLocaleDateString()}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

