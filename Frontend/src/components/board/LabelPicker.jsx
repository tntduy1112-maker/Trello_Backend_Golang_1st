import { useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Check, Pencil, Plus, X } from 'lucide-react';
import labelService from '../../services/label.service';
import { fetchBoard } from '../../redux/slices/boardSlice';

const defaultColors = [
  '#61bd4f', '#f2d600', '#ff9f1a', '#eb5a46', '#c377e0',
  '#0079bf', '#00c2e0', '#51e898', '#ff78cb', '#344563',
];

export default function LabelPicker({ boardLabels, cardLabels, onToggle, onClose }) {
  const dispatch = useDispatch();
  const { currentBoard } = useSelector((state) => state.board);
  const [isCreating, setIsCreating] = useState(false);
  const [editingLabel, setEditingLabel] = useState(null);
  const [name, setName] = useState('');
  const [color, setColor] = useState(defaultColors[0]);

  const isLabelAssigned = (labelId) => cardLabels.some((l) => l.id === labelId);

  const handleCreate = async () => {
    if (!color) return;
    try {
      await labelService.create(currentBoard.id, { name: name || null, color });
      dispatch(fetchBoard(currentBoard.id));
      setIsCreating(false);
      setName('');
      setColor(defaultColors[0]);
    } catch (error) {
      console.error('Failed to create label:', error);
    }
  };

  const handleUpdate = async () => {
    if (!editingLabel || !color) return;
    try {
      await labelService.update(editingLabel.id, { name: name || null, color });
      dispatch(fetchBoard(currentBoard.id));
      setEditingLabel(null);
      setName('');
      setColor(defaultColors[0]);
    } catch (error) {
      console.error('Failed to update label:', error);
    }
  };

  const handleDelete = async (labelId) => {
    if (!confirm('Are you sure you want to delete this label?')) return;
    try {
      await labelService.delete(labelId);
      dispatch(fetchBoard(currentBoard.id));
    } catch (error) {
      console.error('Failed to delete label:', error);
    }
  };

  const startEdit = (label) => {
    setEditingLabel(label);
    setName(label.name || '');
    setColor(label.color);
    setIsCreating(false);
  };

  const renderForm = () => (
    <div className="p-3 border-t">
      <h5 className="font-semibold text-sm mb-2">
        {editingLabel ? 'Edit label' : 'Create label'}
      </h5>
      <input
        type="text"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="Label name (optional)"
        className="w-full px-2 py-1.5 border rounded text-sm mb-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
      <div className="grid grid-cols-5 gap-1 mb-3">
        {defaultColors.map((c) => (
          <button
            key={c}
            onClick={() => setColor(c)}
            className={`h-8 rounded ${color === c ? 'ring-2 ring-offset-1 ring-blue-500' : ''}`}
            style={{ backgroundColor: c }}
          />
        ))}
      </div>
      <div className="flex gap-2">
        <button
          onClick={editingLabel ? handleUpdate : handleCreate}
          className="px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
        >
          {editingLabel ? 'Save' : 'Create'}
        </button>
        {editingLabel && (
          <button
            onClick={() => handleDelete(editingLabel.id)}
            className="px-3 py-1.5 bg-red-600 text-white text-sm rounded hover:bg-red-700"
          >
            Delete
          </button>
        )}
        <button
          onClick={() => {
            setIsCreating(false);
            setEditingLabel(null);
            setName('');
            setColor(defaultColors[0]);
          }}
          className="px-3 py-1.5 text-gray-600 hover:bg-gray-100 rounded text-sm"
        >
          Cancel
        </button>
      </div>
    </div>
  );

  return (
    <>
      <div className="fixed inset-0 z-10" onClick={onClose} />
      <div className="absolute left-0 top-full mt-1 bg-white rounded-lg shadow-lg border z-20 w-72">
        <div className="p-3">
          <h5 className="font-semibold text-sm mb-2">Labels</h5>
          <div className="space-y-1">
            {boardLabels.map((label) => (
              <div key={label.id} className="flex items-center gap-1">
                <button
                  onClick={() => onToggle(label.id, isLabelAssigned(label.id))}
                  className="flex-1 flex items-center gap-2 px-3 py-2 rounded text-sm text-white font-medium hover:opacity-90"
                  style={{ backgroundColor: label.color }}
                >
                  {isLabelAssigned(label.id) && <Check size={16} />}
                  {label.name || ' '}
                </button>
                <button
                  onClick={() => startEdit(label)}
                  className="p-2 hover:bg-gray-100 rounded"
                >
                  <Pencil size={14} className="text-gray-500" />
                </button>
              </div>
            ))}
          </div>
          {!isCreating && !editingLabel && (
            <button
              onClick={() => setIsCreating(true)}
              className="flex items-center gap-2 w-full mt-2 px-3 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded"
            >
              <Plus size={16} />
              Create a new label
            </button>
          )}
        </div>
        {(isCreating || editingLabel) && renderForm()}
      </div>
    </>
  );
}
