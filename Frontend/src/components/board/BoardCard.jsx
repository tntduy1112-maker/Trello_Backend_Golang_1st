import { useDispatch } from 'react-redux';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Calendar, CheckSquare, MessageSquare, Paperclip, User } from 'lucide-react';
import { setOpenCardId } from '../../redux/slices/boardSlice';

const priorityColors = {
  none: '',
  low: 'border-l-4 border-l-green-500',
  medium: 'border-l-4 border-l-yellow-500',
  high: 'border-l-4 border-l-red-500',
};

export default function BoardCard({ card, listId, canEdit = true }) {
  const dispatch = useDispatch();

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: card.id,
    data: { type: 'card', card, listId },
    disabled: !canEdit,
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const handleClick = () => {
    dispatch(setOpenCardId(card.id));
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return null;
    const date = new Date(dateStr);
    const now = new Date();
    const isOverdue = date < now && !card.is_completed;
    const month = date.toLocaleString('en-US', { month: 'short' });
    const day = date.getDate();
    return { text: `${month} ${day}`, isOverdue };
  };

  const dueDate = formatDate(card.due_date);

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      onClick={handleClick}
      className={`bg-white rounded-lg p-2 shadow-sm hover:shadow-md cursor-pointer group ${
        priorityColors[card.priority] || ''
      }`}
    >
      {card.labels && card.labels.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-2">
          {card.labels.map((label) => (
            <span
              key={label.id}
              className="h-2 w-10 rounded-full"
              style={{ backgroundColor: label.color }}
              title={label.name}
            />
          ))}
        </div>
      )}

      <p className="text-sm text-gray-900">{card.title}</p>

      <div className="flex flex-wrap items-center gap-2 mt-2 text-xs text-gray-500">
        {dueDate && (
          <span
            className={`flex items-center gap-1 px-1.5 py-0.5 rounded ${
              card.is_completed
                ? 'bg-green-100 text-green-700'
                : dueDate.isOverdue
                ? 'bg-red-100 text-red-700'
                : 'bg-gray-100'
            }`}
          >
            <Calendar size={12} />
            {dueDate.text}
          </span>
        )}

        {card.description && (
          <span className="flex items-center gap-1" title="Has description">
            <MessageSquare size={12} />
          </span>
        )}

        {card.checklists_progress && (
          <span className="flex items-center gap-1">
            <CheckSquare size={12} />
            {card.checklists_progress.completed}/{card.checklists_progress.total}
          </span>
        )}

        {card.attachments_count > 0 && (
          <span className="flex items-center gap-1">
            <Paperclip size={12} />
            {card.attachments_count}
          </span>
        )}

        {card.assignee && (
          <span className="ml-auto">
            {card.assignee.avatar_url ? (
              <img
                src={card.assignee.avatar_url}
                alt={card.assignee.full_name}
                className="w-6 h-6 rounded-full"
              />
            ) : (
              <span className="flex items-center justify-center w-6 h-6 bg-gray-300 rounded-full text-xs font-medium">
                {card.assignee.full_name?.charAt(0).toUpperCase()}
              </span>
            )}
          </span>
        )}
      </div>
    </div>
  );
}
