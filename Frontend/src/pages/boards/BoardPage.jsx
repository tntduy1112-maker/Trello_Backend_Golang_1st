import { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { fetchBoard } from '../../redux/slices/boardSlice';

export default function BoardPage() {
  const { boardId } = useParams();
  const dispatch = useDispatch();
  const { currentBoard, lists, isLoading } = useSelector((state) => state.board);

  useEffect(() => {
    dispatch(fetchBoard(boardId));
  }, [dispatch, boardId]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-trello-blue"></div>
      </div>
    );
  }

  return (
    <div
      className="min-h-[calc(100vh-48px)] -m-6 p-6"
      style={{ backgroundColor: currentBoard?.background_color || '#0079bf' }}
    >
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold text-white">{currentBoard?.title || 'Board'}</h1>
      </div>

      <div className="flex gap-4 overflow-x-auto pb-4">
        {lists.map((list) => (
          <div key={list.id} className="bg-gray-100 rounded-lg p-3 w-72 flex-shrink-0">
            <h3 className="font-semibold text-gray-900 mb-2">{list.title}</h3>
            <div className="space-y-2">
              {list.cards?.map((card) => (
                <div key={card.id} className="bg-white rounded p-2 shadow-sm hover:shadow cursor-pointer">
                  <p className="text-sm text-gray-900">{card.title}</p>
                </div>
              ))}
            </div>
            <button className="mt-2 text-sm text-gray-500 hover:text-gray-700">
              + Add a card
            </button>
          </div>
        ))}

        <div className="bg-white/30 rounded-lg p-3 w-72 flex-shrink-0">
          <button className="text-white hover:bg-white/20 w-full text-left p-2 rounded">
            + Add another list
          </button>
        </div>
      </div>
    </div>
  );
}
