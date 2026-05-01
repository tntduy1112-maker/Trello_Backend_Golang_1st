import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { forgotPassword, clearError } from '../../redux/slices/authSlice';

export default function ForgotPasswordPage() {
  const dispatch = useDispatch();
  const { isLoading, error } = useSelector((state) => state.auth);

  const [email, setEmail] = useState('');
  const [submitted, setSubmitted] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    const result = await dispatch(forgotPassword(email));
    if (forgotPassword.fulfilled.match(result)) {
      setSubmitted(true);
    }
  };

  if (submitted) {
    return (
      <div className="text-center">
        <h1 className="text-2xl font-bold text-gray-900 mb-4">Check your email</h1>
        <p className="text-gray-600 mb-6">
          If an account exists for <strong>{email}</strong>, you will receive a password reset link.
        </p>
        <Link to="/login" className="link">
          Back to login
        </Link>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-center text-gray-900 mb-2">Forgot password?</h1>
      <p className="text-center text-gray-600 mb-6">
        Enter your email and we&apos;ll send you a reset link.
      </p>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-md text-sm">
          {error.message}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="email" className="label">
            Email
          </label>
          <input
            type="email"
            id="email"
            value={email}
            onChange={(e) => {
              setEmail(e.target.value);
              if (error) dispatch(clearError());
            }}
            required
            className="input"
            placeholder="Enter your email"
          />
        </div>

        <button type="submit" disabled={isLoading} className="btn btn-primary w-full">
          {isLoading ? 'Sending...' : 'Send reset link'}
        </button>
      </form>

      <hr className="my-6" />

      <p className="text-center text-sm text-gray-600">
        <Link to="/login" className="link">
          Back to login
        </Link>
      </p>
    </div>
  );
}
