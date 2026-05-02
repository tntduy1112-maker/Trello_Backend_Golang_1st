import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { Crown, ShieldCheck, Shield, Eye, Loader2, UserPlus, Link2, Copy, Check, LayoutDashboard, ChevronDown, Lock } from 'lucide-react';
import workspaceService from '../../services/workspace.service';
import boardService from '../../services/board.service';

export default function WorkspaceSettingsPage() {
  const { slug } = useParams();
  const navigate = useNavigate();
  const { user } = useSelector((state) => state.auth);

  const [boardMembers, setBoardMembers] = useState([]);
  const [workspace, setWorkspace] = useState(null);
  const [boards, setBoards] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [openDropdownId, setOpenDropdownId] = useState(null);
  const [updatingRole, setUpdatingRole] = useState(null);

  // Invite form state
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState('member');
  const [inviteBoardId, setInviteBoardId] = useState('');
  const [isInviting, setIsInviting] = useState(false);
  const [invitation, setInvitation] = useState(null);
  const [copied, setCopied] = useState(false);

  const isWorkspaceOwner = workspace?.is_workspace_owner;

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [workspaceRes, boardsRes, boardMembersRes] = await Promise.all([
          workspaceService.getBySlug(slug),
          boardService.getByWorkspace(slug),
          workspaceService.getBoardMembers(slug),
        ]);
        setWorkspace(workspaceRes.data.data);
        setBoards(boardsRes.data.data || []);
        setBoardMembers(boardMembersRes.data.data || []);
      } catch (err) {
        setError(err.response?.data?.error?.message || 'Failed to load workspace settings');
      } finally {
        setIsLoading(false);
      }
    };
    fetchData();
  }, [slug]);

  const handleUpdateBoardRole = async (boardId, targetUserId, membershipId, newRole) => {
    setUpdatingRole(membershipId);
    setOpenDropdownId(null);
    setError('');

    try {
      await workspaceService.updateBoardMemberRole(slug, boardId, targetUserId, newRole);
      setBoardMembers(boardMembers.map(member => ({
        ...member,
        boards: member.boards.map(board =>
          board.membership_id === membershipId ? { ...board, role: newRole } : board
        ),
      })));
    } catch (err) {
      setError(err.response?.data?.error?.message || 'Failed to update role');
    } finally {
      setUpdatingRole(null);
    }
  };

  const handleInviteSubmit = async (e) => {
    e.preventDefault();
    if (!inviteBoardId || !inviteEmail) return;

    setIsInviting(true);
    setError('');
    try {
      const response = await boardService.inviteMember(inviteBoardId, inviteEmail, inviteRole);
      setInvitation(response.data.data);
    } catch (err) {
      setError(err.response?.data?.error?.message || 'Failed to create invitation');
    } finally {
      setIsInviting(false);
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
    setInviteEmail('');
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

  const getRoleIcon = (role) => {
    switch (role) {
      case 'owner': return <Crown className="w-3 h-3" />;
      case 'admin': return <ShieldCheck className="w-3 h-3" />;
      case 'member': return <Shield className="w-3 h-3" />;
      case 'viewer': return <Eye className="w-3 h-3" />;
      default: return <Shield className="w-3 h-3" />;
    }
  };

  const getBoardRoleBadgeColor = (role) => {
    switch (role) {
      case 'owner': return 'bg-yellow-50 text-yellow-700 border-yellow-200';
      case 'admin': return 'bg-blue-50 text-blue-700 border-blue-200';
      case 'member': return 'bg-green-50 text-green-700 border-green-200';
      case 'viewer': return 'bg-gray-50 text-gray-600 border-gray-200';
      default: return 'bg-gray-50 text-gray-600 border-gray-200';
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
      </div>
    );
  }

  // Access restriction - only workspace owners can view settings
  if (!isWorkspaceOwner) {
    return (
      <div className="max-w-4xl">
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-8 text-center">
          <Lock className="w-12 h-12 text-yellow-500 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-yellow-800 mb-2">Access Restricted</h2>
          <p className="text-yellow-700 mb-4">
            Only workspace owners can access the settings page.
          </p>
          <p className="text-sm text-yellow-600">
            You need to be an owner of at least one board in this workspace to manage settings.
          </p>
          <button
            onClick={() => navigate(`/workspaces/${slug}`)}
            className="mt-4 px-4 py-2 bg-yellow-600 text-white rounded-lg hover:bg-yellow-700"
          >
            Go to Workspace
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl">
      <h1 className="text-2xl font-bold text-gray-900 mb-2">Workspace Settings</h1>
      <p className="text-gray-600 mb-8">{workspace?.name || slug}</p>

      {error && (
        <div className="mb-4 p-3 bg-red-50 text-red-700 rounded-lg">
          {error}
          <button onClick={() => setError('')} className="ml-2 underline">Dismiss</button>
        </div>
      )}

      {/* Board Members Section */}
      <section className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200 bg-gray-50">
          <div className="flex items-center gap-2">
            <LayoutDashboard className="w-5 h-5 text-gray-600" />
            <h2 className="text-lg font-semibold text-gray-900">Members</h2>
            <span className="text-sm text-gray-500">({boardMembers.length} users)</span>
          </div>
        </div>

        <div className="divide-y divide-gray-100">
          {boardMembers.map((member) => {
            const isCurrentUser = member.user?.id === user?.id;
            const isMemberOwner = member.user?.id === workspace?.owner?.id;

            return (
              <div key={member.user?.id} className="px-6 py-4">
                <div className="flex items-start gap-4">
                  <div className="w-10 h-10 rounded-full bg-blue-100 flex items-center justify-center flex-shrink-0">
                    {member.user?.avatar_url ? (
                      <img
                        src={member.user.avatar_url}
                        alt={member.user.full_name}
                        className="w-10 h-10 rounded-full object-cover"
                      />
                    ) : (
                      <span className="text-blue-600 font-semibold">
                        {member.user?.full_name?.charAt(0)?.toUpperCase() || '?'}
                      </span>
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-2">
                      <span className="font-medium text-gray-900">
                        {member.user?.full_name || 'Unknown'}
                      </span>
                      {isCurrentUser && (
                        <span className="text-xs text-gray-500">(you)</span>
                      )}
                      {isMemberOwner && (
                        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                          <Crown className="w-3 h-3" />
                          Workspace Creator
                        </span>
                      )}
                      <span className="text-sm text-gray-500">{member.user?.email}</span>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      {member.boards?.map((board) => {
                        const dropdownId = `${member.user?.id}-${board.id}`;
                        const isUpdating = updatingRole === board.membership_id;
                        const cannotEditSelf = isCurrentUser && board.role === 'owner';

                        return (
                          <div key={board.id} className="relative">
                            {isUpdating ? (
                              <div className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium border ${getBoardRoleBadgeColor(board.role)}`}>
                                <Loader2 className="w-3 h-3 animate-spin" />
                                <span>{board.title}</span>
                              </div>
                            ) : cannotEditSelf ? (
                              <div
                                className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium border cursor-not-allowed ${getBoardRoleBadgeColor(board.role)}`}
                                title="You cannot change your own owner role"
                              >
                                {getRoleIcon(board.role)}
                                <span>{board.title}</span>
                                <span className="opacity-70">({board.role})</span>
                              </div>
                            ) : (
                              <button
                                onClick={() => setOpenDropdownId(openDropdownId === dropdownId ? null : dropdownId)}
                                className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium border hover:ring-2 hover:ring-blue-200 ${getBoardRoleBadgeColor(board.role)}`}
                              >
                                {getRoleIcon(board.role)}
                                <span>{board.title}</span>
                                <span className="opacity-70">({board.role})</span>
                                <ChevronDown className="w-3 h-3 ml-1" />
                              </button>
                            )}

                            {openDropdownId === dropdownId && (
                              <>
                                <div
                                  className="fixed inset-0 z-10"
                                  onClick={() => setOpenDropdownId(null)}
                                />
                                <div className="absolute left-0 mt-1 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-1 z-20">
                                  {['owner', 'admin', 'member', 'viewer'].map((role) => (
                                    <button
                                      key={role}
                                      onClick={() => handleUpdateBoardRole(board.id, member.user.id, board.membership_id, role)}
                                      disabled={board.role === role}
                                      className="w-full px-3 py-2 text-left text-sm hover:bg-gray-50 flex items-center gap-2 disabled:opacity-50 disabled:bg-gray-50"
                                    >
                                      {getRoleIcon(role)}
                                      <span className="capitalize">{role}</span>
                                      {board.role === role && (
                                        <span className="ml-auto text-xs text-gray-400">Current</span>
                                      )}
                                    </button>
                                  ))}
                                </div>
                              </>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {boardMembers.length === 0 && (
          <div className="px-6 py-8 text-center text-gray-500">
            No members found
          </div>
        )}
      </section>

      {/* Invite to Board Section */}
      {boards.length > 0 && (
        <section className="mt-6 bg-white rounded-lg border border-gray-200 overflow-hidden">
          <div className="px-6 py-4 border-b border-gray-200 bg-gray-50">
            <div className="flex items-center gap-2">
              <UserPlus className="w-5 h-5 text-gray-600" />
              <h2 className="text-lg font-semibold text-gray-900">Invite to Board</h2>
            </div>
          </div>

          <div className="p-6">
            {invitation ? (
              <div className="space-y-4">
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

                <button
                  onClick={handleCreateAnother}
                  className="w-full px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50"
                >
                  Create Another Invitation
                </button>
              </div>
            ) : (
              <form onSubmit={handleInviteSubmit} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Board
                  </label>
                  <select
                    value={inviteBoardId}
                    onChange={(e) => setInviteBoardId(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    required
                  >
                    <option value="">Select a board</option>
                    {boards.map((board) => (
                      <option key={board.id} value={board.id}>
                        {board.title}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Email address
                  </label>
                  <input
                    type="email"
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
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
                    value={inviteRole}
                    onChange={(e) => setInviteRole(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  >
                    <option value="member">Member - Edit cards</option>
                    <option value="admin">Admin - Manage board, edit lists, invite members</option>
                    <option value="viewer">Viewer - Read-only access</option>
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

                <button
                  type="submit"
                  disabled={isInviting || !inviteBoardId || !inviteEmail}
                  className="w-full flex items-center justify-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isInviting ? (
                    <Loader2 className="w-4 h-4 animate-spin" />
                  ) : (
                    <UserPlus className="w-4 h-4" />
                  )}
                  Create Invitation Link
                </button>
              </form>
            )}
          </div>
        </section>
      )}

      {/* Role Legend */}
      <div className="mt-6 p-4 bg-gray-50 rounded-lg">
        <h3 className="text-sm font-medium text-gray-700 mb-3">Board Role Permissions</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 text-sm">
          <div className="flex items-start gap-2">
            <Crown className="w-4 h-4 text-yellow-500 mt-0.5 flex-shrink-0" />
            <div>
              <span className="font-medium text-gray-900">Owner</span>
              <p className="text-gray-500">Full board control, manage workspace members, invite to any board</p>
            </div>
          </div>
          <div className="flex items-start gap-2">
            <ShieldCheck className="w-4 h-4 text-blue-500 mt-0.5 flex-shrink-0" />
            <div>
              <span className="font-medium text-gray-900">Admin</span>
              <p className="text-gray-500">Manage board, edit lists, invite members to their board</p>
            </div>
          </div>
          <div className="flex items-start gap-2">
            <Shield className="w-4 h-4 text-green-500 mt-0.5 flex-shrink-0" />
            <div>
              <span className="font-medium text-gray-900">Member</span>
              <p className="text-gray-500">Edit cards</p>
            </div>
          </div>
          <div className="flex items-start gap-2">
            <Eye className="w-4 h-4 text-gray-400 mt-0.5 flex-shrink-0" />
            <div>
              <span className="font-medium text-gray-900">Viewer</span>
              <p className="text-gray-500">Read-only access</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
