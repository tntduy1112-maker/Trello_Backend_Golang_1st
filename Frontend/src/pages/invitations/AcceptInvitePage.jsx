import { useParams, useNavigate } from 'react-router-dom';

export default function AcceptInvitePage() {
  const { token } = useParams();
  const navigate = useNavigate();

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100">
      <div className="bg-white rounded-lg shadow-lg p-8 max-w-md w-full text-center">
        <h1 className="text-2xl font-bold text-gray-900 mb-4">Board Invitation</h1>
        <p className="text-gray-600 mb-6">
          You have been invited to join a board.
        </p>
        <p className="text-sm text-gray-500 mb-4">Token: {token}</p>
        <button onClick={() => navigate('/login')} className="btn btn-primary w-full">
          Log in to accept
        </button>
      </div>
    </div>
  );
}
