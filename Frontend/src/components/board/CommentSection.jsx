import { useState, useEffect, useRef } from 'react';
import { useSelector } from 'react-redux';
import { MessageSquare, Send, Edit2, Trash2, CornerDownRight } from 'lucide-react';
import commentService from '../../services/comment.service';
import MentionInput, { formatMentionsForDisplay } from '../ui/MentionInput';

export default function CommentSection({ cardId, canEdit = true, members = [] }) {
  const { user } = useSelector((state) => state.auth);
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState('');
  const [editingId, setEditingId] = useState(null);
  const [editContent, setEditContent] = useState('');
  const [replyingTo, setReplyingTo] = useState(null);
  const [replyContent, setReplyContent] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [highlightedCommentId, setHighlightedCommentId] = useState(null);
  const commentRefs = useRef({});

  useEffect(() => {
    loadComments();
  }, [cardId]);

  // Scroll to specific comment if coming from notification
  useEffect(() => {
    if (!isLoading && comments.length > 0) {
      const scrollToId = sessionStorage.getItem('scrollToCommentId');
      if (scrollToId) {
        sessionStorage.removeItem('scrollToCommentId');

        // Small delay to ensure DOM is ready
        setTimeout(() => {
          const commentEl = commentRefs.current[scrollToId];
          if (commentEl) {
            commentEl.scrollIntoView({ behavior: 'smooth', block: 'center' });
            setHighlightedCommentId(scrollToId);
            // Remove highlight after 3 seconds
            setTimeout(() => setHighlightedCommentId(null), 3000);
          }
        }, 100);
      }
    }
  }, [isLoading, comments]);

  const loadComments = async () => {
    try {
      const res = await commentService.list(cardId);
      setComments(res.data.data || []);
    } catch (error) {
      console.error('Failed to load comments:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    if (e) e.preventDefault();
    if (!newComment.trim()) return;

    try {
      const res = await commentService.create(cardId, newComment.trim());
      setComments([res.data.data, ...comments]);
      setNewComment('');
    } catch (error) {
      console.error('Failed to create comment:', error);
    }
  };

  const handleReply = async (parentId) => {
    if (!replyContent.trim()) return;

    try {
      const res = await commentService.create(cardId, replyContent.trim(), parentId);
      setComments(
        comments.map((c) =>
          c.id === parentId ? { ...c, replies: [...(c.replies || []), res.data.data] } : c
        )
      );
      setReplyingTo(null);
      setReplyContent('');
    } catch (error) {
      console.error('Failed to reply:', error);
    }
  };

  const handleUpdate = async (commentId) => {
    if (!editContent.trim()) return;

    try {
      await commentService.update(commentId, editContent.trim());
      setComments(
        comments.map((c) => (c.id === commentId ? { ...c, content: editContent.trim() } : c))
      );
      setEditingId(null);
      setEditContent('');
    } catch (error) {
      console.error('Failed to update comment:', error);
    }
  };

  const handleDelete = async (commentId) => {
    if (!confirm('Delete this comment?')) return;

    try {
      await commentService.delete(commentId);
      setComments(comments.filter((c) => c.id !== commentId));
    } catch (error) {
      console.error('Failed to delete comment:', error);
    }
  };

  const formatDate = (dateStr) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
    });
  };

  const renderCommentContent = (content) => {
    if (!content) return null;
    const mentionRegex = /@\[([^\]]+)\]\([^)]+\)/g;
    const parts = [];
    let lastIndex = 0;
    let match;

    while ((match = mentionRegex.exec(content)) !== null) {
      if (match.index > lastIndex) {
        parts.push(<span key={`text-${lastIndex}`}>{content.slice(lastIndex, match.index)}</span>);
      }
      parts.push(
        <span key={`mention-${match.index}`} className="text-blue-600 font-medium bg-blue-50 px-1 rounded">
          @{match[1]}
        </span>
      );
      lastIndex = match.index + match[0].length;
    }

    if (lastIndex < content.length) {
      parts.push(<span key={`text-${lastIndex}`}>{content.slice(lastIndex)}</span>);
    }

    return parts.length > 0 ? parts : content;
  };

  const renderComment = (comment, isReply = false) => (
    <div
      key={comment.id}
      ref={(el) => (commentRefs.current[comment.id] = el)}
      className={`flex gap-3 ${isReply ? 'ml-8 mt-3' : ''} ${
        highlightedCommentId === comment.id
          ? 'bg-yellow-100 -mx-2 px-2 py-1 rounded-lg ring-2 ring-yellow-400 transition-all duration-300'
          : ''
      }`}
    >
      <div className="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center text-white text-sm font-medium flex-shrink-0">
        {comment.author?.full_name?.charAt(0).toUpperCase() || '?'}
      </div>
      <div className="flex-1">
        <div className="flex items-center gap-2">
          <span className="font-semibold text-sm">{comment.author?.full_name}</span>
          <span className="text-xs text-gray-500">{formatDate(comment.created_at)}</span>
        </div>

        {editingId === comment.id ? (
          <div className="mt-1">
            <textarea
              value={editContent}
              onChange={(e) => setEditContent(e.target.value)}
              className="w-full p-2 border rounded text-sm resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
              rows={2}
              autoFocus
            />
            <div className="flex gap-2 mt-1">
              <button
                onClick={() => handleUpdate(comment.id)}
                className="px-2 py-1 bg-blue-600 text-white text-xs rounded hover:bg-blue-700"
              >
                Save
              </button>
              <button
                onClick={() => {
                  setEditingId(null);
                  setEditContent('');
                }}
                className="px-2 py-1 text-gray-600 text-xs hover:bg-gray-100 rounded"
              >
                Cancel
              </button>
            </div>
          </div>
        ) : (
          <>
            <p className="text-sm mt-1 bg-white p-2 rounded shadow-sm">
                    {renderCommentContent(comment.content)}
                  </p>
            <div className="flex items-center gap-3 mt-1">
              {canEdit && !isReply && (
                <button
                  onClick={() => setReplyingTo(comment.id)}
                  className="text-xs text-gray-500 hover:text-gray-700 flex items-center gap-1"
                >
                  <CornerDownRight size={12} /> Reply
                </button>
              )}
              {canEdit && comment.author?.id === user?.id && (
                <>
                  <button
                    onClick={() => {
                      setEditingId(comment.id);
                      setEditContent(comment.content);
                    }}
                    className="text-xs text-gray-500 hover:text-gray-700 flex items-center gap-1"
                  >
                    <Edit2 size={12} /> Edit
                  </button>
                  <button
                    onClick={() => handleDelete(comment.id)}
                    className="text-xs text-red-500 hover:text-red-700 flex items-center gap-1"
                  >
                    <Trash2 size={12} /> Delete
                  </button>
                </>
              )}
            </div>
          </>
        )}

        {replyingTo === comment.id && (
          <div className="mt-2 flex gap-2">
            <input
              type="text"
              value={replyContent}
              onChange={(e) => setReplyContent(e.target.value)}
              placeholder="Write a reply..."
              className="flex-1 px-3 py-1.5 border rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              autoFocus
            />
            <button
              onClick={() => handleReply(comment.id)}
              className="px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
            >
              Reply
            </button>
            <button
              onClick={() => {
                setReplyingTo(null);
                setReplyContent('');
              }}
              className="px-3 py-1.5 text-gray-600 text-sm hover:bg-gray-100 rounded"
            >
              Cancel
            </button>
          </div>
        )}

        {comment.replies?.map((reply) => renderComment(reply, true))}
      </div>
    </div>
  );

  return (
    <div>
      <div className="flex items-center gap-2 mb-3">
        <MessageSquare size={20} className="text-gray-500" />
        <h4 className="font-semibold">Comments</h4>
      </div>

      {canEdit && (
        <form onSubmit={handleSubmit} className="flex gap-2 mb-4">
          <div className="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center text-white text-sm font-medium flex-shrink-0">
            {user?.full_name?.charAt(0).toUpperCase() || '?'}
          </div>
          <MentionInput
            value={newComment}
            onChange={setNewComment}
            onSubmit={handleSubmit}
            members={members}
            placeholder="Write a comment... Use @ to mention"
          />
          <button
            type="submit"
            disabled={!newComment.trim()}
            className="px-3 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Send size={16} />
          </button>
        </form>
      )}

      {isLoading ? (
        <div className="text-center py-4 text-gray-500 text-sm">Loading comments...</div>
      ) : comments.length === 0 ? (
        <div className="text-center py-4 text-gray-500 text-sm">No comments yet</div>
      ) : (
        <div className="space-y-4">{comments.map((comment) => renderComment(comment))}</div>
      )}
    </div>
  );
}
