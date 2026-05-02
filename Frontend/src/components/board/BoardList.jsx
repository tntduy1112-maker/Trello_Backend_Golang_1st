import { useState } from 'react';
import { useDispatch } from 'react-redux';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { MoreHorizontal, Plus, X } from 'lucide-react';
import { updateList, deleteList, createCard } from '../../redux/slices/boardSlice';
import BoardCard from './BoardCard';

export default function BoardList({ list, cards = [], canEdit = true, canManage = false }) {
  const dispatch = useDispatch();
  const [isEditing, setIsEditing] = useState(false);
  const [title, setTitle] = useState(list.title);
  const [showAddCard, setShowAddCard] = useState(false);
  const [newCardTitle, setNewCardTitle] = useState('');
  const [showMenu, setShowMenu] = useState(false);

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: list.id,
    data: { type: 'list', list },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const handleTitleSubmit = () => {
    if (title.trim() && title !== list.title) {
      dispatch(updateList({ listId: list.id, data: { title: title.trim() } }));
    }
    setIsEditing(false);
  };

  const handleAddCard = (e) => {
    e.preventDefault();
    if (newCardTitle.trim()) {
      dispatch(createCard({ listId: list.id, title: newCardTitle.trim() }));
      setNewCardTitle('');
      setShowAddCard(false);
    }
  };

  const handleArchiveList = () => {
    if (confirm('Are you sure you want to archive this list?')) {
      dispatch(deleteList(list.id));
    }
    setShowMenu(false);
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className="bg-gray-100 rounded-xl w-72 flex-shrink-0 flex flex-col max-h-[calc(100vh-140px)]"
    >
      <div
        {...attributes}
        {...listeners}
        className="p-3 cursor-grab active:cursor-grabbing"
      >
        <div className="flex items-center justify-between">
          {isEditing && canManage ? (
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              onBlur={handleTitleSubmit}
              onKeyDown={(e) => e.key === 'Enter' && handleTitleSubmit()}
              className="flex-1 px-2 py-1 text-sm font-semibold bg-white border-2 border-blue-500 rounded focus:outline-none"
              autoFocus
            />
          ) : (
            <h3
              onClick={() => canManage && setIsEditing(true)}
              className={`font-semibold text-gray-900 px-2 py-1 rounded flex-1 ${canManage ? 'cursor-pointer hover:bg-gray-200' : ''}`}
            >
              {list.title}
            </h3>
          )}
          {canManage && (
            <div className="relative">
              <button
                onClick={() => setShowMenu(!showMenu)}
                className="p-1 rounded hover:bg-gray-200 text-gray-500"
              >
                <MoreHorizontal size={16} />
              </button>
              {showMenu && (
                <>
                  <div className="fixed inset-0 z-10" onClick={() => setShowMenu(false)} />
                  <div className="absolute right-0 top-8 bg-white rounded-lg shadow-lg border py-1 z-20 w-48">
                    <button
                      onClick={handleArchiveList}
                      className="w-full text-left px-4 py-2 text-sm hover:bg-gray-100 text-red-600"
                    >
                      Archive list
                    </button>
                  </div>
                </>
              )}
            </div>
          )}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto px-3 space-y-2">
        <SortableContext items={cards.map((c) => c.id)} strategy={verticalListSortingStrategy}>
          {cards.map((card) => (
            <BoardCard key={card.id} card={card} listId={list.id} canEdit={canEdit} />
          ))}
        </SortableContext>
      </div>

      {canEdit && (
        <div className="p-3">
          {showAddCard ? (
            <form onSubmit={handleAddCard}>
              <textarea
                value={newCardTitle}
                onChange={(e) => setNewCardTitle(e.target.value)}
                placeholder="Enter a title for this card..."
                className="w-full p-2 text-sm border rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
                rows={3}
                autoFocus
              />
              <div className="flex items-center gap-2 mt-2">
                <button
                  type="submit"
                  className="px-3 py-1.5 bg-blue-600 text-white text-sm font-medium rounded hover:bg-blue-700"
                >
                  Add card
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowAddCard(false);
                    setNewCardTitle('');
                  }}
                  className="p-1.5 text-gray-500 hover:text-gray-700"
                >
                  <X size={20} />
                </button>
              </div>
            </form>
          ) : (
            <button
              onClick={() => setShowAddCard(true)}
              className="flex items-center gap-1 w-full p-2 text-sm text-gray-600 hover:bg-gray-200 rounded-lg"
            >
              <Plus size={16} />
              Add a card
            </button>
          )}
        </div>
      )}
    </div>
  );
}
