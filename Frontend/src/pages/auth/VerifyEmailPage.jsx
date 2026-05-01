import { useState } from 'react';
import { useLocation, useNavigate, Link } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { verifyEmail, clearError } from '../../redux/slices/authSlice';
import authService from '../../services/auth.service';

export default function VerifyEmailPage() {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const location = useLocation();
  const { isLoading, error } = useSelector((state) => state.auth);

  const email = location.state?.email || '';
  const [otp, setOtp] = useState('');
  const [resendLoading, setResendLoading] = useState(false);
  const [resendMessage, setResendMessage] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    const result = await dispatch(verifyEmail({ email, otp }));
    if (verifyEmail.fulfilled.match(result)) {
      navigate('/login', { state: { message: 'Email verified! Please log in.' } });
    }
  };

  const handleResend = async () => {
    setResendLoading(true);
    setResendMessage('');
    try {
      await authService.resendVerification(email);
      setResendMessage('Verification code sent!');
    } catch (err) {
      setResendMessage(err.response?.data?.error?.message || 'Failed to resend code');
    } finally {
      setResendLoading(false);
    }
  };

  return (
    <div>
      <h1 className="text-2xl font-bold text-center text-gray-900 mb-2">Verify your email</h1>
      <p className="text-center text-gray-600 mb-6">
        We sent a verification code to <strong>{email}</strong>
      </p>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-md text-sm">
          {error.message}
        </div>
      )}

      {resendMessage && (
        <div className="mb-4 p-3 bg-green-50 border border-green-200 text-green-700 rounded-md text-sm">
          {resendMessage}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="otp" className="label">
            Verification Code
          </label>
          <input
            type="text"
            id="otp"
            value={otp}
            onChange={(e) => {
              setOtp(e.target.value);
              if (error) dispatch(clearError());
            }}
            required
            maxLength={6}
            className="input text-center text-2xl tracking-widest"
            placeholder="000000"
          />
        </div>

        <button type="submit" disabled={isLoading} className="btn btn-primary w-full">
          {isLoading ? 'Verifying...' : 'Verify Email'}
        </button>
      </form>

      <div className="mt-4 text-center">
        <button
          onClick={handleResend}
          disabled={resendLoading}
          className="text-sm link"
        >
          {resendLoading ? 'Sending...' : "Didn't receive the code? Resend"}
        </button>
      </div>

      <hr className="my-6" />

      <p className="text-center text-sm text-gray-600">
        <Link to="/login" className="link">
          Back to login
        </Link>
      </p>
    </div>
  );
}
