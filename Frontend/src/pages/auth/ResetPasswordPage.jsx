import { useState } from 'react';
import { Link, useSearchParams, useNavigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { resetPassword, clearError } from '../../redux/slices/authSlice';

export default function ResetPasswordPage() {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { isLoading, error } = useSelector((state) => state.auth);

  const token = searchParams.get('token') || '';
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [validationError, setValidationError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (password !== confirmPassword) {
      setValidationError('Passwords do not match');
      return;
    }
    const result = await dispatch(resetPassword({ token, password }));
    if (resetPassword.fulfilled.match(result)) {
      navigate('/login', { state: { message: 'Password reset successful! Please log in.' } });
    }
  };

  return (
    <div>
      <h1 className="text-2xl font-bold text-center text-gray-900 mb-6">Reset your password</h1>

      {(error || validationError) && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-md text-sm">
          {error?.message || validationError}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="password" className="label">New Password</label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={(e) => { setPassword(e.target.value); setValidationError(''); if (error) dispatch(clearError()); }}
            required
            minLength={6}
            className="input"
            placeholder="Enter new password"
          />
        </div>

        <div>
          <label htmlFor="confirmPassword" className="label">Confirm Password</label>
          <input
            type="password"
            id="confirmPassword"
            value={confirmPassword}
            onChange={(e) => { setConfirmPassword(e.target.value); setValidationError(''); }}
            required
            className="input"
            placeholder="Confirm new password"
          />
        </div>

        <button type="submit" disabled={isLoading} className="btn btn-primary w-full">
          {isLoading ? 'Resetting...' : 'Reset Password'}
        </button>
      </form>

      <hr className="my-6" />
      <p className="text-center text-sm text-gray-600">
        <Link to="/login" className="link">Back to login</Link>
      </p>
    </div>
  );
}
