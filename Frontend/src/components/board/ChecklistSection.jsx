import { useState, useEffect } from 'react';
import { CheckSquare, Plus, Trash2, X } from 'lucide-react';
import checklistService from '../../services/checklist.service';

export default function ChecklistSection({ cardId, canEdit = true }) {
  const [checklists, setChecklists] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [newChecklistTitle, setNewChecklistTitle] = useState('');
  const [showAddChecklist, setShowAddChecklist] = useState(false);
  const [addingItemTo, setAddingItemTo] = useState(null);
  const [newItemTitle, setNewItemTitle] = useState('');

  useEffect(() => {
    loadChecklists();
  }, [cardId]);

  const loadChecklists = async () => {
    try {
      const res = await checklistService.list(cardId);
      setChecklists(res.data.data || []);
    } catch (error) {
      console.error('Failed to load checklists:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateChecklist = async (e) => {
    e.preventDefault();
    if (!newChecklistTitle.trim()) return;

    try {
      const res = await checklistService.create(cardId, newChecklistTitle.trim());
      setChecklists([...checklists, { ...res.data.data, items: [] }]);
      setNewChecklistTitle('');
      setShowAddChecklist(false);
    } catch (error) {
      console.error('Failed to create checklist:', error);
    }
  };

  const handleDeleteChecklist = async (checklistId) => {
    if (!confirm('Delete this checklist?')) return;

    try {
      await checklistService.delete(checklistId);
      setChecklists(checklists.filter((c) => c.id !== checklistId));
    } catch (error) {
      console.error('Failed to delete checklist:', error);
    }
  };

  const handleAddItem = async (checklistId) => {
    if (!newItemTitle.trim()) return;

    try {
      const res = await checklistService.createItem(checklistId, newItemTitle.trim());
      setChecklists(
        checklists.map((c) =>
          c.id === checklistId ? { ...c, items: [...(c.items || []), res.data.data] } : c
        )
      );
      setNewItemTitle('');
      setAddingItemTo(null);
    } catch (error) {
      console.error('Failed to add item:', error);
    }
  };

  const handleToggleItem = async (checklistId, itemId) => {
    try {
      const res = await checklistService.toggleItem(itemId);
      setChecklists(
        checklists.map((c) =>
          c.id === checklistId
            ? {
                ...c,
                items: c.items.map((item) => (item.id === itemId ? res.data.data : item)),
              }
            : c
        )
      );
    } catch (error) {
      console.error('Failed to toggle item:', error);
    }
  };

  const handleDeleteItem = async (checklistId, itemId) => {
    try {
      await checklistService.deleteItem(itemId);
      setChecklists(
        checklists.map((c) =>
          c.id === checklistId ? { ...c, items: c.items.filter((item) => item.id !== itemId) } : c
        )
      );
    } catch (error) {
      console.error('Failed to delete item:', error);
    }
  };

  const getProgress = (checklist) => {
    const items = checklist.items || [];
    if (items.length === 0) return 0;
    const completed = items.filter((item) => item.is_completed).length;
    return Math.round((completed / items.length) * 100);
  };

  if (isLoading) {
    return <div className="text-center py-4 text-gray-500 text-sm">Loading checklists...</div>;
  }

  return (
    <div className="space-y-4">
      {checklists.map((checklist) => (
        <div key={checklist.id} className="bg-white rounded-lg p-3 shadow-sm">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <CheckSquare size={18} className="text-gray-500" />
              <h5 className="font-semibold text-sm">{checklist.title}</h5>
            </div>
            {canEdit && (
              <button
                onClick={() => handleDeleteChecklist(checklist.id)}
                className="p-1 text-gray-400 hover:text-red-500 rounded"
              >
                <Trash2 size={14} />
              </button>
            )}
          </div>

          <div className="mb-2">
            <div className="flex items-center gap-2 text-xs text-gray-500 mb-1">
              <span>{getProgress(checklist)}%</span>
            </div>
            <div className="h-1.5 bg-gray-200 rounded-full overflow-hidden">
              <div
                className="h-full bg-green-500 transition-all"
                style={{ width: `${getProgress(checklist)}%` }}
              />
            </div>
          </div>

          <div className="space-y-1">
            {(checklist.items || []).map((item) => (
              <div key={item.id} className="flex items-center gap-2 group">
                <input
                  type="checkbox"
                  checked={item.is_completed}
                  onChange={() => canEdit && handleToggleItem(checklist.id, item.id)}
                  disabled={!canEdit}
                  className="rounded text-blue-600 focus:ring-blue-500"
                />
                <span
                  className={`flex-1 text-sm ${item.is_completed ? 'line-through text-gray-400' : ''}`}
                >
                  {item.title}
                </span>
                {canEdit && (
                  <button
                    onClick={() => handleDeleteItem(checklist.id, item.id)}
                    className="p-1 text-gray-300 hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity"
                  >
                    <X size={14} />
                  </button>
                )}
              </div>
            ))}
          </div>

          {canEdit && (addingItemTo === checklist.id ? (
            <div className="mt-2 flex gap-2">
              <input
                type="text"
                value={newItemTitle}
                onChange={(e) => setNewItemTitle(e.target.value)}
                placeholder="Add an item..."
                className="flex-1 px-2 py-1 border rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleAddItem(checklist.id);
                  if (e.key === 'Escape') {
                    setAddingItemTo(null);
                    setNewItemTitle('');
                  }
                }}
              />
              <button
                onClick={() => handleAddItem(checklist.id)}
                className="px-2 py-1 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
              >
                Add
              </button>
              <button
                onClick={() => {
                  setAddingItemTo(null);
                  setNewItemTitle('');
                }}
                className="px-2 py-1 text-gray-600 text-sm hover:bg-gray-100 rounded"
              >
                Cancel
              </button>
            </div>
          ) : (
            <button
              onClick={() => setAddingItemTo(checklist.id)}
              className="mt-2 flex items-center gap-1 text-sm text-gray-500 hover:text-gray-700"
            >
              <Plus size={14} /> Add an item
            </button>
          ))}
        </div>
      ))}

      {canEdit && (showAddChecklist ? (
        <form onSubmit={handleCreateChecklist} className="bg-white rounded-lg p-3 shadow-sm">
          <input
            type="text"
            value={newChecklistTitle}
            onChange={(e) => setNewChecklistTitle(e.target.value)}
            placeholder="Checklist title..."
            className="w-full px-2 py-1 border rounded text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 mb-2"
            autoFocus
          />
          <div className="flex gap-2">
            <button
              type="submit"
              className="px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
            >
              Add
            </button>
            <button
              type="button"
              onClick={() => {
                setShowAddChecklist(false);
                setNewChecklistTitle('');
              }}
              className="px-3 py-1.5 text-gray-600 text-sm hover:bg-gray-100 rounded"
            >
              Cancel
            </button>
          </div>
        </form>
      ) : (
        <button
          onClick={() => setShowAddChecklist(true)}
          className="flex items-center gap-2 w-full px-3 py-2 bg-gray-200 rounded hover:bg-gray-300 text-sm"
        >
          <CheckSquare size={16} />
          Add checklist
        </button>
      ))}
    </div>
  );
}
