import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { Bell, Check, CheckCheck, MessageSquare, User, AtSign, CreditCard, Calendar } from 'lucide-react';
import { fetchNotifications, markAsRead, markAllAsRead } from '../../redux/slices/notificationSlice';
import { setOpenCardId } from '../../redux/slices/boardSlice';

const notificationIcons = {
  mentioned: AtSign,
  comment_added: MessageSquare,
  comment_reply: MessageSquare,
  card_assigned: User,
  card_due_soon: Calendar,
  card_overdue: Calendar,
  board_invitation: CreditCard,
  member_added_to_board: User,
};

const notificationColors = {
  mentioned: 'bg-purple-100 text-purple-600',
  comment_added: 'bg-blue-100 text-blue-600',
  comment_reply: 'bg-blue-100 text-blue-600',
  card_assigned: 'bg-green-100 text-green-600',
  card_due_soon: 'bg-yellow-100 text-yellow-600',
  card_overdue: 'bg-red-100 text-red-600',
  board_invitation: 'bg-indigo-100 text-indigo-600',
  member_added_to_board: 'bg-teal-100 text-teal-600',
};

const formatCommentPreview = (text) => {
  if (!text) return '';
  return text.replace(/@\[([^\]]+)\]\([^)]+\)/g, '@$1');
};

export default function NotificationDropdown({ onClose }) {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { notifications, isLoading } = useSelector((state) => state.notification);
  const { currentBoard } = useSelector((state) => state.board);

  useEffect(() => {
    dispatch(fetchNotifications());
  }, [dispatch]);

  const handleMarkAsRead = (e, id) => {
    e.stopPropagation();
    dispatch(markAsRead(id));
  };

  const handleMarkAllAsRead = () => {
    dispatch(markAllAsRead());
  };

  const handleNotificationClick = (notification) => {
    // Mark as read
    if (!notification.is_read) {
      dispatch(markAsRead(notification.id));
    }

    // Navigate to the card if card_id exists
    if (notification.card_id) {
      // Build navigation context for the card modal
      const navContext = {
        tab: notification.type === 'mentioned' || notification.type === 'comment_added' || notification.type === 'comment_reply' ? 'comments' : 'details',
        commentId: notification.metadata?.comment_id || null,
      };

      // If we're on the same board, just open the card modal
      if (currentBoard?.id === notification.board_id) {
        dispatch(setOpenCardId({ cardId: notification.card_id, context: navContext }));
      } else {
        // Navigate to the board first, then open card
        sessionStorage.setItem('openCardAfterNav', JSON.stringify({
          cardId: notification.card_id,
          context: navContext,
        }));
        navigate(`/board/${notification.board_id}`);
      }
      onClose();
    } else if (notification.board_id) {
      navigate(`/board/${notification.board_id}`);
      onClose();
    }
  };

  const formatTime = (dateString) => {
    const date = new Date(dateString);
    const now = new Date();
    const diff = now - date;
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 1) return 'Just now';
    if (minutes < 60) return `${minutes}m ago`;
    if (hours < 24) return `${hours}h ago`;
    return `${days}d ago`;
  };

  const getIcon = (type) => {
    return notificationIcons[type] || Bell;
  };

  const getColorClass = (type) => {
    return notificationColors[type] || 'bg-gray-100 text-gray-600';
  };

  return (
    <div className="absolute right-0 top-12 w-96 bg-white rounded-lg shadow-lg border border-gray-200 z-50">
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-100">
        <h3 className="font-semibold text-gray-900">Notifications</h3>
        <button
          onClick={handleMarkAllAsRead}
          className="text-sm text-trello-blue hover:underline flex items-center gap-1"
        >
          <CheckCheck className="w-4 h-4" />
          Mark all read
        </button>
      </div>

      <div className="max-h-[480px] overflow-y-auto">
        {isLoading ? (
          <div className="p-4 text-center text-gray-500">Loading...</div>
        ) : notifications.length === 0 ? (
          <div className="p-8 text-center">
            <Bell className="w-12 h-12 mx-auto text-gray-300 mb-2" />
            <p className="text-gray-500">No notifications yet</p>
          </div>
        ) : (
          notifications.map((notification) => {
            const Icon = getIcon(notification.type);
            const colorClass = getColorClass(notification.type);
            const hasLink = notification.card_id || notification.board_id;

            return (
              <div
                key={notification.id}
                onClick={() => hasLink && handleNotificationClick(notification)}
                className={`px-4 py-3 border-b border-gray-50 transition-colors ${
                  !notification.is_read ? 'bg-blue-50' : ''
                } ${hasLink ? 'cursor-pointer hover:bg-gray-50' : ''}`}
              >
                <div className="flex items-start gap-3">
                  <div className={`w-9 h-9 rounded-full flex items-center justify-center flex-shrink-0 ${colorClass}`}>
                    <Icon className="w-4 h-4" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900">{notification.title}</p>
                    {notification.message && (
                      <p className="text-sm text-gray-600 mt-0.5 line-clamp-2">{notification.message}</p>
                    )}
                    {notification.metadata?.comment_preview && (
                      <p className="text-xs text-gray-500 mt-1 italic line-clamp-1 bg-gray-100 px-2 py-1 rounded">
                        "{formatCommentPreview(notification.metadata.comment_preview)}"
                      </p>
                    )}
                    <div className="flex items-center gap-2 mt-1">
                      <p className="text-xs text-gray-400">{formatTime(notification.created_at)}</p>
                      {hasLink && (
                        <span className="text-xs text-blue-500">Click to view</span>
                      )}
                    </div>
                  </div>
                  {!notification.is_read && (
                    <button
                      onClick={(e) => handleMarkAsRead(e, notification.id)}
                      className="p-1 text-gray-400 hover:text-trello-blue flex-shrink-0"
                      title="Mark as read"
                    >
                      <Check className="w-4 h-4" />
                    </button>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
