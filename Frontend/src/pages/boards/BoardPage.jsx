import { useEffect, useCallback, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import {
  DndContext,
  DragOverlay,
  closestCorners,
  PointerSensor,
  useSensor,
  useSensors,
} from '@dnd-kit/core';
import { SortableContext, horizontalListSortingStrategy } from '@dnd-kit/sortable';
import { UserPlus } from 'lucide-react';
import {
  fetchBoard,
  moveCard,
  moveList,
  optimisticMoveCard,
  optimisticMoveList,
  setOpenCardId,
} from '../../redux/slices/boardSlice';
import BoardList from '../../components/board/BoardList';
import BoardCard from '../../components/board/BoardCard';
import AddListForm from '../../components/board/AddListForm';
import CardDetailModal from '../../components/board/CardDetailModal';
import InviteMemberModal from '../../components/board/InviteMemberModal';

export default function BoardPage() {
  const { boardId } = useParams();
  const dispatch = useDispatch();
  const { currentBoard, lists, isLoading, openCardId } = useSelector((state) => state.board);
  const [activeItem, setActiveItem] = useState(null);
  const [showInviteModal, setShowInviteModal] = useState(false);

  const myRole = currentBoard?.my_role;
  const canManage = myRole === 'owner' || myRole === 'admin';
  const canEdit = canManage || myRole === 'member';
  const isViewer = myRole === 'viewer';

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 5 },
    })
  );

  useEffect(() => {
    dispatch(fetchBoard(boardId));
  }, [dispatch, boardId]);

  // Check if we should open a card after navigation (from notification click)
  useEffect(() => {
    const navDataStr = sessionStorage.getItem('openCardAfterNav');
    if (navDataStr && currentBoard?.id === boardId) {
      sessionStorage.removeItem('openCardAfterNav');
      try {
        const navData = JSON.parse(navDataStr);
        dispatch(setOpenCardId({ cardId: navData.cardId, context: navData.context }));
      } catch (e) {
        // Legacy format - just a card ID string
        dispatch(setOpenCardId(navDataStr));
      }
    }
  }, [dispatch, boardId, currentBoard]);

  const findListByCardId = useCallback(
    (cardId) => {
      return lists.find((list) => list.cards?.some((card) => card.id === cardId));
    },
    [lists]
  );

  const calculatePosition = (items, overIndex, activeIndex) => {
    if (items.length === 0) return 65536;

    if (overIndex === 0) {
      return items[0].position / 2;
    }

    if (overIndex >= items.length) {
      return items[items.length - 1].position + 65536;
    }

    const beforeItem = items[overIndex - 1];
    const afterItem = items[overIndex];

    if (activeIndex !== undefined && activeIndex < overIndex) {
      return (afterItem.position + (items[overIndex + 1]?.position || afterItem.position + 65536)) / 2;
    }

    return (beforeItem.position + afterItem.position) / 2;
  };

  const handleDragStart = (event) => {
    const { active } = event;
    setActiveItem(active.data.current);
  };

  const handleDragEnd = (event) => {
    const { active, over } = event;
    setActiveItem(null);

    if (!over) return;

    const activeData = active.data.current;
    const overData = over.data.current;

    if (activeData.type === 'list' && overData?.type === 'list') {
      if (active.id === over.id) return;

      const oldIndex = lists.findIndex((l) => l.id === active.id);
      const newIndex = lists.findIndex((l) => l.id === over.id);
      const newPosition = calculatePosition(lists, newIndex, oldIndex);

      dispatch(optimisticMoveList({ listId: active.id, newPosition }));
      dispatch(moveList({ listId: active.id, position: newPosition }));
    }

    if (activeData.type === 'card') {
      const activeList = findListByCardId(active.id);
      let targetListId;
      let overIndex;

      if (overData?.type === 'card') {
        const overList = findListByCardId(over.id);
        targetListId = overList.id;
        overIndex = overList.cards.findIndex((c) => c.id === over.id);
      } else if (overData?.type === 'list') {
        targetListId = over.id;
        const targetList = lists.find((l) => l.id === targetListId);
        overIndex = targetList.cards?.length || 0;
      } else {
        return;
      }

      const targetList = lists.find((l) => l.id === targetListId);
      const cards = targetList.cards || [];
      const activeIndex = activeList.id === targetListId
        ? cards.findIndex((c) => c.id === active.id)
        : undefined;

      const newPosition = calculatePosition(cards, overIndex, activeIndex);

      dispatch(
        optimisticMoveCard({
          cardId: active.id,
          fromListId: activeList.id,
          toListId: targetListId,
          newPosition,
        })
      );

      dispatch(
        moveCard({
          cardId: active.id,
          listId: targetListId,
          position: newPosition,
        })
      );
    }
  };

  const handleDragOver = (event) => {
    const { active, over } = event;
    if (!over) return;

    const activeData = active.data.current;
    const overData = over.data.current;

    if (activeData.type !== 'card') return;

    const activeList = findListByCardId(active.id);
    let targetListId;

    if (overData?.type === 'card') {
      const overList = findListByCardId(over.id);
      targetListId = overList?.id;
    } else if (overData?.type === 'list') {
      targetListId = over.id;
    }

    if (!targetListId || activeList?.id === targetListId) return;

    const targetList = lists.find((l) => l.id === targetListId);
    const overIndex = overData?.type === 'card'
      ? targetList.cards?.findIndex((c) => c.id === over.id) ?? 0
      : targetList.cards?.length ?? 0;

    const newPosition = calculatePosition(targetList.cards || [], overIndex);

    dispatch(
      optimisticMoveCard({
        cardId: active.id,
        fromListId: activeList.id,
        toListId: targetListId,
        newPosition,
      })
    );
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
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
        {canManage && (
          <button
            onClick={() => setShowInviteModal(true)}
            className="flex items-center gap-2 px-3 py-1.5 bg-white/20 hover:bg-white/30 text-white rounded-lg text-sm font-medium transition-colors"
          >
            <UserPlus className="w-4 h-4" />
            Invite
          </button>
        )}
      </div>

      <DndContext
        sensors={sensors}
        collisionDetection={closestCorners}
        onDragStart={handleDragStart}
        onDragOver={handleDragOver}
        onDragEnd={handleDragEnd}
      >
        <div className="flex gap-4 overflow-x-auto pb-4 items-start">
          <SortableContext items={lists.filter((l) => l.id).map((l) => l.id)} strategy={horizontalListSortingStrategy}>
            {lists.filter((list) => list.id).map((list) => (
              <BoardList key={list.id} list={list} cards={list.cards || []} canEdit={canEdit} canManage={canManage} />
            ))}
          </SortableContext>

          {canManage && <AddListForm boardId={boardId} />}
        </div>

        <DragOverlay>
          {activeItem?.type === 'list' && (
            <div className="bg-gray-100 rounded-xl w-72 p-3 shadow-xl opacity-90">
              <h3 className="font-semibold text-gray-900">{activeItem.list.title}</h3>
            </div>
          )}
          {activeItem?.type === 'card' && (
            <div className="bg-white rounded-lg p-2 shadow-xl w-64 opacity-90">
              <p className="text-sm text-gray-900">{activeItem.card.title}</p>
            </div>
          )}
        </DragOverlay>
      </DndContext>

      {openCardId && <CardDetailModal />}
      {showInviteModal && (
        <InviteMemberModal
          boardId={boardId}
          onClose={() => setShowInviteModal(false)}
        />
      )}
    </div>
  );
}
