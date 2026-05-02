import { useState } from 'react';
import { X, UserPlus, Loader2, Copy, Check, Link2 } from 'lucide-react';
import boardService from '../../services/board.service';

export default function InviteMemberModal({ boardId, onClose }) {
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('member');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [invitation, setInvitation] = useState(null);
  const [copied, setCopied] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setInvitation(null);
    setIsLoading(true);

    try {
      const response = await boardService.inviteMember(boardId, email, role);
      setInvitation(response.data.data);
    } catch (err) {
      setError(err.response?.data?.error?.message || 'Failed to create invitation');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCopyLink = async () => {
    if (invitation?.invite_url) {
      await navigator.clipboard.writeText(invitation.invite_url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleCreateAnother = () => {
    setInvitation(null);
    setEmail('');
    setCopied(false);
  };

  const formatExpiry = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-md mx-4">
        <div className="flex items-center justify-between px-4 py-3 border-b">
          <h2 className="text-lg font-semibold text-gray-900">
            {invitation ? 'Invitation Created' : 'Invite to Board'}
          </h2>
          <button onClick={onClose} className="p-1 hover:bg-gray-100 rounded">
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        {invitation ? (
          <div className="p-4 space-y-4">
            <div className="p-4 bg-green-50 rounded-lg">
              <div className="flex items-center gap-2 text-green-700 mb-2">
                <Check className="w-5 h-5" />
                <span className="font-medium">Invitation created!</span>
              </div>
              <p className="text-sm text-green-600">
                Share this link with <strong>{invitation.invitee_email}</strong>
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Invitation Link
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  readOnly
                  value={invitation.invite_url}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 text-sm"
                />
                <button
                  onClick={handleCopyLink}
                  className="px-3 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-1"
                >
                  {copied ? (
                    <>
                      <Check className="w-4 h-4" />
                      Copied
                    </>
                  ) : (
                    <>
                      <Copy className="w-4 h-4" />
                      Copy
                    </>
                  )}
                </button>
              </div>
            </div>

            <div className="text-sm text-gray-500 space-y-1">
              <p><strong>Role:</strong> {invitation.role}</p>
              <p><strong>Expires:</strong> {formatExpiry(invitation.expires_at)}</p>
            </div>

            <div className="flex gap-2">
              <button
                onClick={handleCreateAnother}
                className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50"
              >
                Create Another
              </button>
              <button
                onClick={onClose}
                className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
              >
                Done
              </button>
            </div>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="p-4 space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Email address
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="Enter email address"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Role
              </label>
              <select
                value={role}
                onChange={(e) => setRole(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="member">Member - Can view and edit cards</option>
                <option value="admin">Admin - Can manage board settings</option>
              </select>
            </div>

            <div className="p-3 bg-gray-50 rounded-lg text-sm text-gray-600">
              <div className="flex items-center gap-2 mb-1">
                <Link2 className="w-4 h-4" />
                <span className="font-medium">Shareable Link</span>
              </div>
              <p>A unique invitation link will be generated. Share it with the invitee to join this board.</p>
              <p className="mt-1 text-gray-500">Link expires in 3 days.</p>
            </div>

            {error && (
              <div className="p-3 bg-red-50 text-red-700 text-sm rounded-lg">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={isLoading || !email}
              className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <UserPlus className="w-4 h-4" />
              )}
              Create Invitation Link
            </button>
          </form>
        )}
      </div>
    </div>
  );
}
