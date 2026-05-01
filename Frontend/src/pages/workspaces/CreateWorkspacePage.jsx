import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { createWorkspace } from '../../redux/slices/workspaceSlice';

export default function CreateWorkspacePage() {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { isLoading } = useSelector((state) => state.workspace);

  const [formData, setFormData] = useState({ name: '', description: '' });
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    const result = await dispatch(createWorkspace(formData));
    if (createWorkspace.fulfilled.match(result)) {
      navigate(`/workspaces/${result.payload.slug}`);
    } else {
      setError(result.payload?.message || 'Failed to create workspace');
    }
  };

  return (
    <div className="max-w-lg mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Create Workspace</h1>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-700 rounded-md text-sm">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="card p-6 space-y-4">
        <div>
          <label htmlFor="name" className="label">Workspace Name</label>
          <input
            type="text"
            id="name"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            required
            className="input"
            placeholder="e.g., My Team"
          />
        </div>

        <div>
          <label htmlFor="description" className="label">Description (optional)</label>
          <textarea
            id="description"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            className="input min-h-[100px]"
            placeholder="What is this workspace for?"
          />
        </div>

        <div className="flex gap-3">
          <button type="button" onClick={() => navigate(-1)} className="btn btn-secondary flex-1">
            Cancel
          </button>
          <button type="submit" disabled={isLoading} className="btn btn-primary flex-1">
            {isLoading ? 'Creating...' : 'Create Workspace'}
          </button>
        </div>
      </form>
    </div>
  );
}
