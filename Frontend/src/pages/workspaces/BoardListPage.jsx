import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { Plus, LayoutGrid } from 'lucide-react';
import { fetchWorkspace } from '../../redux/slices/workspaceSlice';
import { fetchBoards, createBoard } from '../../redux/slices/boardSlice';

export default function BoardListPage() {
  const { slug } = useParams();
  const dispatch = useDispatch();
  const { currentWorkspace } = useSelector((state) => state.workspace);
  const { boards, isLoading } = useSelector((state) => state.board);

  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newBoardTitle, setNewBoardTitle] = useState('');
  const [newBoardColor, setNewBoardColor] = useState('#0079bf');

  useEffect(() => {
    dispatch(fetchWorkspace(slug));
    dispatch(fetchBoards(slug));
  }, [dispatch, slug]);

  const handleCreateBoard = async (e) => {
    e.preventDefault();
    const result = await dispatch(createBoard({
      workspaceSlug: slug,
      data: { title: newBoardTitle, background_color: newBoardColor },
    }));
    if (createBoard.fulfilled.match(result)) {
      setShowCreateModal(false);
      setNewBoardTitle('');
    }
  };

  const colors = ['#0079bf', '#d29034', '#519839', '#b04632', '#89609e', '#cd5a91', '#4bbf6b', '#00aecc'];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">
          {currentWorkspace?.name || 'Workspace'}
        </h1>
        <button onClick={() => setShowCreateModal(true)} className="btn btn-primary flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Create Board
        </button>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-trello-blue"></div>
        </div>
      ) : boards.length === 0 ? (
        <div className="text-center py-12">
          <LayoutGrid className="w-16 h-16 mx-auto text-gray-300 mb-4" />
          <h2 className="text-xl font-medium text-gray-600 mb-2">No boards yet</h2>
          <p className="text-gray-500 mb-4">Create your first board to start organizing</p>
          <button onClick={() => setShowCreateModal(true)} className="btn btn-primary">
            Create Board
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {boards.map((board) => (
            <Link
              key={board.id}
              to={`/board/${board.id}`}
              className="h-24 rounded-lg p-4 text-white font-semibold hover:opacity-90 transition-opacity"
              style={{ backgroundColor: board.background_color || '#0079bf' }}
            >
              {board.title}
            </Link>
          ))}
        </div>
      )}

      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-bold mb-4">Create Board</h2>
            <form onSubmit={handleCreateBoard} className="space-y-4">
              <div>
                <label className="label">Board Title</label>
                <input
                  type="text"
                  value={newBoardTitle}
                  onChange={(e) => setNewBoardTitle(e.target.value)}
                  required
                  className="input"
                  placeholder="Enter board title"
                />
              </div>
              <div>
                <label className="label">Background Color</label>
                <div className="flex gap-2">
                  {colors.map((color) => (
                    <button
                      key={color}
                      type="button"
                      onClick={() => setNewBoardColor(color)}
                      className={`w-8 h-8 rounded ${newBoardColor === color ? 'ring-2 ring-offset-2 ring-trello-blue' : ''}`}
                      style={{ backgroundColor: color }}
                    />
                  ))}
                </div>
              </div>
              <div className="flex gap-3">
                <button type="button" onClick={() => setShowCreateModal(false)} className="btn btn-secondary flex-1">
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary flex-1">
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
