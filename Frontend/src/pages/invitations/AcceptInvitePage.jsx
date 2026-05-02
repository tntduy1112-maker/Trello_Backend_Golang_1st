import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useSelector, useDispatch } from 'react-redux';
import { Loader2, CheckCircle, XCircle, Clock, Mail, Eye, EyeOff, UserPlus } from 'lucide-react';
import invitationService from '../../services/invitation.service';
import { setCredentials } from '../../redux/slices/authSlice';

export default function AcceptInvitePage() {
  const { token } = useParams();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { isAuthenticated, user } = useSelector((state) => state.auth);

  const [invitation, setInvitation] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [isAccepting, setIsAccepting] = useState(false);
  const [accepted, setAccepted] = useState(false);

  // New user setup state
  const [fullName, setFullName] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    const fetchInvitation = async () => {
      try {
        const response = await invitationService.getByToken(token);
        setInvitation(response.data.data);
      } catch (err) {
        setError(err.response?.data?.error?.message || 'Invalid or expired invitation');
      } finally {
        setIsLoading(false);
      }
    };

    fetchInvitation();
  }, [token]);

  const handleAccept = async () => {
    setIsAccepting(true);
    setError('');

    try {
      await invitationService.accept(token);
      setAccepted(true);
      setTimeout(() => {
        navigate(`/board/${invitation.board.id}`);
      }, 2000);
    } catch (err) {
      setError(err.response?.data?.error?.message || 'Failed to accept invitation');
    } finally {
      setIsAccepting(false);
    }
  };

  const handleAcceptWithPassword = async (e) => {
    e.preventDefault();
    setError('');

    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }

    setIsAccepting(true);

    try {
      const response = await invitationService.acceptWithPassword(token, {
        full_name: fullName,
        password: password,
      });

      const { user: newUser, access_token, refresh_token, board_id } = response.data.data;

      localStorage.setItem('accessToken', access_token);
      localStorage.setItem('refreshToken', refresh_token);

      dispatch(setCredentials({
        user: newUser,
        accessToken: access_token,
        refreshToken: refresh_token,
      }));

      setAccepted(true);
      setTimeout(() => {
        navigate(`/board/${board_id}`);
      }, 2000);
    } catch (err) {
      setError(err.response?.data?.error?.message || 'Failed to create account');
    } finally {
      setIsAccepting(false);
    }
  };

  const formatExpiry = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const isExpired = invitation && new Date(invitation.expires_at) < new Date();
  const emailMismatch = isAuthenticated && user && invitation && user.email !== invitation.invitee_email;

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="text-center">
          <Loader2 className="w-12 h-12 animate-spin text-blue-600 mx-auto mb-4" />
          <p className="text-gray-600">Loading invitation...</p>
        </div>
      </div>
    );
  }

  if (error && !invitation) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="bg-white rounded-lg shadow-lg p-8 max-w-md w-full text-center">
          <XCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Invalid Invitation</h1>
          <p className="text-gray-600 mb-6">{error}</p>
          <Link
            to="/"
            className="inline-block px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Go to Home
          </Link>
        </div>
      </div>
    );
  }

  if (accepted) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="bg-white rounded-lg shadow-lg p-8 max-w-md w-full text-center">
          <CheckCircle className="w-16 h-16 text-green-500 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Welcome to the Board!</h1>
          <p className="text-gray-600 mb-4">
            You've successfully joined <strong>{invitation.board.title}</strong>
          </p>
          <p className="text-sm text-gray-500">Redirecting to board...</p>
        </div>
      </div>
    );
  }

  if (isExpired) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-100">
        <div className="bg-white rounded-lg shadow-lg p-8 max-w-md w-full text-center">
          <Clock className="w-16 h-16 text-orange-500 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Invitation Expired</h1>
          <p className="text-gray-600 mb-6">
            This invitation has expired. Please ask the board owner to send a new invitation.
          </p>
          <Link
            to="/"
            className="inline-block px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Go to Home
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100">
      <div className="bg-white rounded-lg shadow-lg p-8 max-w-md w-full">
        <div className="text-center mb-6">
          <h1 className="text-2xl font-bold text-gray-900 mb-2">You're Invited!</h1>
          <p className="text-gray-600">
            {invitation.inviter?.full_name || 'Someone'} invited you to join a board
          </p>
        </div>

        <div className="bg-blue-50 rounded-lg p-4 mb-6">
          <h2 className="text-lg font-semibold text-blue-900 mb-2">
            {invitation.board?.title}
          </h2>
          <div className="text-sm text-blue-700 space-y-1">
            <p><strong>Role:</strong> {invitation.role === 'admin' ? 'Admin' : 'Member'}</p>
            <p><strong>Expires:</strong> {formatExpiry(invitation.expires_at)}</p>
          </div>
        </div>

        <div className="bg-gray-50 rounded-lg p-4 mb-6">
          <div className="flex items-center gap-2 text-gray-700 mb-1">
            <Mail className="w-4 h-4" />
            <span className="text-sm font-medium">This invitation is for:</span>
          </div>
          <p className="text-gray-900 font-medium">{invitation.invitee_email}</p>
        </div>

        {error && (
          <div className="p-3 bg-red-50 text-red-700 text-sm rounded-lg mb-4">
            {error}
          </div>
        )}

        {emailMismatch && (
          <div className="p-3 bg-yellow-50 text-yellow-800 text-sm rounded-lg mb-4">
            <p className="font-medium mb-1">Email mismatch</p>
            <p>
              You're logged in as <strong>{user.email}</strong>, but this invitation was sent to{' '}
              <strong>{invitation.invitee_email}</strong>.
            </p>
            <p className="mt-2">Please log in with the correct email to accept this invitation.</p>
          </div>
        )}

        {isAuthenticated ? (
          <button
            onClick={handleAccept}
            disabled={isAccepting || emailMismatch}
            className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
          >
            {isAccepting ? (
              <>
                <Loader2 className="w-5 h-5 animate-spin" />
                Accepting...
              </>
            ) : (
              <>
                <CheckCircle className="w-5 h-5" />
                Accept Invitation
              </>
            )}
          </button>
        ) : invitation.is_new_user ? (
          <form onSubmit={handleAcceptWithPassword} className="space-y-4">
            <div className="p-3 bg-green-50 text-green-800 text-sm rounded-lg">
              <p className="font-medium mb-1">
                <UserPlus className="w-4 h-4 inline mr-1" />
                Create your account
              </p>
              <p>Set up your password to join the board.</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Full Name
              </label>
              <input
                type="text"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                placeholder="Enter your full name"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                required
                minLength={2}
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Create a password"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 pr-10"
                  required
                  minLength={8}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700"
                >
                  {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                </button>
              </div>
              <p className="text-xs text-gray-500 mt-1">At least 8 characters</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Confirm Password
              </label>
              <input
                type={showPassword ? 'text' : 'password'}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="Confirm your password"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                required
              />
            </div>

            <button
              type="submit"
              disabled={isAccepting || !fullName || !password || !confirmPassword}
              className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
            >
              {isAccepting ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  Creating account...
                </>
              ) : (
                <>
                  <UserPlus className="w-5 h-5" />
                  Create Account & Join Board
                </>
              )}
            </button>
          </form>
        ) : (
          <div className="space-y-3">
            <Link
              to={`/login?redirect=/invite/${token}`}
              className="w-full flex items-center justify-center px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium"
            >
              Log in to Accept
            </Link>
            <Link
              to={`/register?email=${encodeURIComponent(invitation.invitee_email)}&redirect=/invite/${token}`}
              className="w-full flex items-center justify-center px-4 py-3 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 font-medium"
            >
              Create Account
            </Link>
          </div>
        )}
      </div>
    </div>
  );
}
