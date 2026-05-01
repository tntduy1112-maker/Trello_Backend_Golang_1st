import { useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { Plus, LayoutGrid } from 'lucide-react';
import { fetchWorkspaces } from '../../redux/slices/workspaceSlice';

export default function WorkspacesPage() {
  const dispatch = useDispatch();
  const { workspaces, isLoading } = useSelector((state) => state.workspace);

  useEffect(() => {
    dispatch(fetchWorkspaces());
  }, [dispatch]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-trello-blue"></div>
      </div>
    );
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Your Workspaces</h1>
        <Link to="/workspaces/create" className="btn btn-primary flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Create Workspace
        </Link>
      </div>

      {workspaces.length === 0 ? (
        <div className="text-center py-12">
          <LayoutGrid className="w-16 h-16 mx-auto text-gray-300 mb-4" />
          <h2 className="text-xl font-medium text-gray-600 mb-2">No workspaces yet</h2>
          <p className="text-gray-500 mb-4">Create your first workspace to get started</p>
          <Link to="/workspaces/create" className="btn btn-primary">
            Create Workspace
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {workspaces.map((workspace) => (
            <Link
              key={workspace.id}
              to={`/workspaces/${workspace.slug}`}
              className="card p-4 hover:shadow-md transition-shadow"
            >
              <div className="flex items-center gap-3">
                <div
                  className="w-12 h-12 rounded flex items-center justify-center text-white text-xl font-bold"
                  style={{ backgroundColor: '#0079bf' }}
                >
                  {workspace.name[0].toUpperCase()}
                </div>
                <div>
                  <h3 className="font-semibold text-gray-900">{workspace.name}</h3>
                  <p className="text-sm text-gray-500">
                    {workspace.boards_count || 0} boards
                  </p>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
