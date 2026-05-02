import { useState, useEffect } from 'react';
import { Activity, ChevronDown } from 'lucide-react';
import activityService from '../../services/activity.service';

const ACTION_MESSAGES = {
  card_created: 'created this card',
  card_moved: 'moved this card',
  card_archived: 'archived this card',
  card_restored: 'restored this card',
  card_assigned: 'assigned this card',
  card_unassigned: 'removed assignment from this card',
  card_completed: 'marked this card as complete',
  card_incompleted: 'marked this card as incomplete',
  label_added: 'added a label to this card',
  label_removed: 'removed a label from this card',
  due_date_set: 'set the due date',
  due_date_removed: 'removed the due date',
  comment_added: 'commented on this card',
  comment_updated: 'updated a comment',
  comment_deleted: 'deleted a comment',
  checklist_created: 'added a checklist',
  checklist_deleted: 'deleted a checklist',
  checklist_item_created: 'added a checklist item',
  checklist_item_completed: 'completed a checklist item',
  checklist_item_uncompleted: 'uncompleted a checklist item',
  checklist_item_deleted: 'deleted a checklist item',
  attachment_added: 'attached a file',
  attachment_deleted: 'deleted an attachment',
  cover_set: 'set the cover',
  cover_removed: 'removed the cover',
};

const getActionMessage = (activity) => {
  const { action, metadata, description } = activity;

  // Use backend description if available (it has more detail)
  if (description) {
    return description;
  }

  let message = ACTION_MESSAGES[action] || action;

  if (metadata) {
    if (action === 'card_moved' && metadata.from_list && metadata.to_list) {
      message = `moved this card from ${metadata.from_list} to ${metadata.to_list}`;
    }
    if (action === 'card_assigned' && metadata.assignee_name) {
      message = `assigned ${metadata.assignee_name} to this card`;
    }
    if (action === 'checklist_created' && metadata.checklist_name) {
      message = `added checklist "${metadata.checklist_name}"`;
    }
    if (action === 'attachment_added' && metadata.attachment_name) {
      message = `attached "${metadata.attachment_name}"`;
    }
  }

  return message;
};

export default function ActivitySection({ cardId, boardId }) {
  const [activities, setActivities] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showAll, setShowAll] = useState(false);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);

  useEffect(() => {
    loadActivities();
  }, [cardId, boardId]);

  const loadActivities = async (loadPage = 1) => {
    try {
      const res = cardId
        ? await activityService.listByCard(cardId, loadPage, 10)
        : await activityService.listByBoard(boardId, loadPage, 10);

      const data = res.data.data || [];
      if (loadPage === 1) {
        setActivities(data);
      } else {
        setActivities((prev) => [...prev, ...data]);
      }
      setHasMore(data.length === 10);
      setPage(loadPage);
    } catch (error) {
      console.error('Failed to load activity:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const loadMore = () => {
    loadActivities(page + 1);
  };

  const formatDate = (dateStr) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;

    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined,
    });
  };

  const displayedActivities = showAll ? activities : activities.slice(0, 5);

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <Activity size={20} className="text-gray-500" />
          <h4 className="font-semibold">Activity</h4>
        </div>
        {activities.length > 5 && (
          <button
            onClick={() => setShowAll(!showAll)}
            className="text-sm text-blue-600 hover:underline"
          >
            {showAll ? 'Show less' : 'Show all'}
          </button>
        )}
      </div>

      {isLoading ? (
        <div className="text-center py-4 text-gray-500 text-sm">Loading activity...</div>
      ) : activities.length === 0 ? (
        <div className="text-center py-4 text-gray-500 text-sm">No activity yet</div>
      ) : (
        <div className="space-y-3">
          {displayedActivities.map((activity) => (
            <div key={activity.id} className="flex gap-3">
              <div className="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center text-gray-600 text-sm font-medium flex-shrink-0">
                {activity.user?.full_name?.charAt(0).toUpperCase() || '?'}
              </div>
              <div className="flex-1">
                <p className="text-sm">
                  <span className="font-semibold">{activity.user?.full_name}</span>{' '}
                  {getActionMessage(activity)}
                </p>
                <span className="text-xs text-gray-500">{formatDate(activity.created_at)}</span>
              </div>
            </div>
          ))}

          {showAll && hasMore && (
            <button
              onClick={loadMore}
              className="flex items-center gap-1 text-sm text-blue-600 hover:underline mx-auto"
            >
              <ChevronDown size={14} /> Load more
            </button>
          )}
        </div>
      )}
    </div>
  );
}
