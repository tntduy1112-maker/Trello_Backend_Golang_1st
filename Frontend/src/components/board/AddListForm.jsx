import { useState } from 'react';
import { useDispatch } from 'react-redux';
import { Plus, X } from 'lucide-react';
import { createList } from '../../redux/slices/boardSlice';

export default function AddListForm({ boardId }) {
  const dispatch = useDispatch();
  const [isAdding, setIsAdding] = useState(false);
  const [title, setTitle] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    if (title.trim()) {
      dispatch(createList({ boardId, title: title.trim() }));
      setTitle('');
      setIsAdding(false);
    }
  };

  if (!isAdding) {
    return (
      <button
        onClick={() => setIsAdding(true)}
        className="flex items-center gap-2 bg-white/30 hover:bg-white/40 rounded-xl p-3 w-72 flex-shrink-0 text-white font-medium transition-colors"
      >
        <Plus size={20} />
        Add another list
      </button>
    );
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="bg-gray-100 rounded-xl p-3 w-72 flex-shrink-0"
    >
      <input
        type="text"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        placeholder="Enter list title..."
        className="w-full px-3 py-2 text-sm border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        autoFocus
      />
      <div className="flex items-center gap-2 mt-2">
        <button
          type="submit"
          className="px-3 py-1.5 bg-blue-600 text-white text-sm font-medium rounded hover:bg-blue-700"
        >
          Add list
        </button>
        <button
          type="button"
          onClick={() => {
            setIsAdding(false);
            setTitle('');
          }}
          className="p-1.5 text-gray-500 hover:text-gray-700"
        >
          <X size={20} />
        </button>
      </div>
    </form>
  );
}
