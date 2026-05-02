import { useState, useEffect, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import {
  X,
  CreditCard,
  AlignLeft,
  Tag,
  User,
  Calendar,
  CheckSquare,
  Paperclip,
  Archive,
  MessageSquare,
  Activity,
} from 'lucide-react';
import { setOpenCardId, updateCard, deleteCard, clearCardNavContext } from '../../redux/slices/boardSlice';
import cardService from '../../services/card.service';
import LabelPicker from './LabelPicker';
import CommentSection from './CommentSection';
import ChecklistSection from './ChecklistSection';
import AttachmentSection from './AttachmentSection';
import ActivitySection from './ActivitySection';

const priorityOptions = [
  { value: 'none', label: 'None', color: 'bg-gray-200' },
  { value: 'low', label: 'Low', color: 'bg-green-500' },
  { value: 'medium', label: 'Medium', color: 'bg-yellow-500' },
  { value: 'high', label: 'High', color: 'bg-red-500' },
];

export default function CardDetailModal() {
  const dispatch = useDispatch();
  const { openCardId, lists, labels: boardLabels, members, currentBoard, cardNavContext } = useSelector((state) => state.board);
  const [card, setCard] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [isEditingDesc, setIsEditingDesc] = useState(false);
  const [showLabelPicker, setShowLabelPicker] = useState(false);
  const [showMemberPicker, setShowMemberPicker] = useState(false);
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [activeTab, setActiveTab] = useState('details');

  // Use ref to capture cardNavContext so we can use it when openCardId changes
  const navContextRef = useRef(null);

  // Keep ref updated with latest cardNavContext
  useEffect(() => {
    if (cardNavContext) {
      navContextRef.current = cardNavContext;
    }
  }, [cardNavContext]);

  const myRole = currentBoard?.my_role;
  const canEdit = myRole === 'owner' || myRole === 'admin' || myRole === 'member';

  useEffect(() => {
    if (openCardId) {
      setIsLoading(true);

      // Use navigation context from ref (set by notification click)
      const ctx = navContextRef.current;

      if (ctx) {
        setActiveTab(ctx.tab || 'details');
        if (ctx.commentId) {
          sessionStorage.setItem('scrollToCommentId', ctx.commentId);
        }
        navContextRef.current = null;
        dispatch(clearCardNavContext());
      } else {
        setActiveTab('details');
      }

      cardService
        .getById(openCardId)
        .then((res) => {
          setCard(res.data);
          setTitle(res.data.title);
          setDescription(res.data.description || '');
        })
        .catch(console.error)
        .finally(() => setIsLoading(false));
    }
  }, [openCardId, dispatch]);

  if (!openCardId) return null;

  const handleClose = () => {
    dispatch(setOpenCardId(null));
    setCard(null);
  };

  const handleTitleBlur = () => {
    if (title.trim() && title !== card.title) {
      dispatch(updateCard({ cardId: card.id, data: { title: title.trim() } }));
      setCard({ ...card, title: title.trim() });
    }
  };

  const handleDescriptionSave = () => {
    if (description !== card.description) {
      dispatch(updateCard({ cardId: card.id, data: { description } }));
      setCard({ ...card, description });
    }
    setIsEditingDesc(false);
  };

  const handlePriorityChange = (priority) => {
    dispatch(updateCard({ cardId: card.id, data: { priority } }));
    setCard({ ...card, priority });
  };

  const handleToggleComplete = async () => {
    try {
      if (card.is_completed) {
        await cardService.markIncomplete(card.id);
      } else {
        await cardService.markComplete(card.id);
      }
      setCard({ ...card, is_completed: !card.is_completed });
    } catch (error) {
      console.error('Failed to toggle complete:', error);
    }
  };

  const handleAssign = async (userId) => {
    try {
      await cardService.assign(card.id, userId);
      const member = members.find((m) => m.user.id === userId);
      setCard({ ...card, assignee: member?.user });
      setShowMemberPicker(false);
    } catch (error) {
      console.error('Failed to assign:', error);
    }
  };

  const handleUnassign = async () => {
    try {
      await cardService.unassign(card.id);
      setCard({ ...card, assignee: null });
    } catch (error) {
      console.error('Failed to unassign:', error);
    }
  };

  const handleDueDateChange = async (e) => {
    const dueDate = e.target.value ? new Date(e.target.value).toISOString() : null;
    dispatch(updateCard({ cardId: card.id, data: { due_date: dueDate } }));
    setCard({ ...card, due_date: dueDate });
    setShowDatePicker(false);
  };

  const handleArchive = () => {
    if (confirm('Are you sure you want to archive this card?')) {
      dispatch(deleteCard(card.id));
      handleClose();
    }
  };

  const handleLabelToggle = async (labelId, isAssigned) => {
    try {
      if (isAssigned) {
        await cardService.removeLabel(card.id, labelId);
        setCard({
          ...card,
          labels: card.labels.filter((l) => l.id !== labelId),
        });
      } else {
        await cardService.addLabel(card.id, labelId);
        const label = boardLabels.find((l) => l.id === labelId);
        setCard({
          ...card,
          labels: [...(card.labels || []), label],
        });
      }
    } catch (error) {
      console.error('Failed to toggle label:', error);
    }
  };

  const currentList = lists.find(
    (l) => l.id === card?.list?.id || l.cards?.some((c) => c.id === card?.id)
  );

  const tabs = [
    { id: 'details', label: 'Details', icon: CreditCard },
    { id: 'checklists', label: 'Checklists', icon: CheckSquare },
    { id: 'attachments', label: 'Attachments', icon: Paperclip },
    { id: 'comments', label: 'Comments', icon: MessageSquare },
    { id: 'activity', label: 'Activity', icon: Activity },
  ];

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-full items-start justify-center p-4 pt-16">
        <div className="fixed inset-0 bg-black/50" onClick={handleClose} />

        <div className="relative bg-gray-100 rounded-xl shadow-xl w-full max-w-3xl max-h-[85vh] overflow-hidden flex flex-col">
          <button
            onClick={handleClose}
            className="absolute right-4 top-4 p-1 rounded-full hover:bg-gray-200 z-10"
          >
            <X size={20} />
          </button>

          {isLoading ? (
            <div className="p-8 text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto" />
            </div>
          ) : card ? (
            <>
              <div className="p-6 pb-0">
                <div className="flex items-start gap-3 mb-4">
                  <CreditCard size={24} className="text-gray-500 mt-1" />
                  <div className="flex-1">
                    {canEdit ? (
                      <input
                        type="text"
                        value={title}
                        onChange={(e) => setTitle(e.target.value)}
                        onBlur={handleTitleBlur}
                        className="text-xl font-semibold bg-transparent w-full focus:bg-white focus:px-2 focus:py-1 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                    ) : (
                      <h2 className="text-xl font-semibold">{title}</h2>
                    )}
                    <p className="text-sm text-gray-500 mt-1">
                      in list <span className="underline">{currentList?.title || card.list?.title}</span>
                    </p>
                  </div>
                </div>

                <div className="flex border-b border-gray-300 -mx-6 px-6">
                  {tabs.map((tab) => (
                    <button
                      key={tab.id}
                      onClick={() => setActiveTab(tab.id)}
                      className={`flex items-center gap-1.5 px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
                        activeTab === tab.id
                          ? 'border-blue-600 text-blue-600'
                          : 'border-transparent text-gray-500 hover:text-gray-700'
                      }`}
                    >
                      <tab.icon size={16} />
                      {tab.label}
                    </button>
                  ))}
                </div>
              </div>

              <div className="flex-1 overflow-y-auto p-6">
                {activeTab === 'details' && (
                  <div className="flex gap-6">
                    <div className="flex-1 space-y-6">
                      {card.labels && card.labels.length > 0 && (
                        <div>
                          <h4 className="text-xs font-semibold text-gray-500 uppercase mb-2">
                            Labels
                          </h4>
                          <div className="flex flex-wrap gap-1">
                            {card.labels.map((label) => (
                              <span
                                key={label.id}
                                className="px-3 py-1 rounded text-sm text-white font-medium"
                                style={{ backgroundColor: label.color }}
                              >
                                {label.name || ' '}
                              </span>
                            ))}
                          </div>
                        </div>
                      )}

                      {card.assignee && (
                        <div>
                          <h4 className="text-xs font-semibold text-gray-500 uppercase mb-2">
                            Assignee
                          </h4>
                          <div className="flex items-center gap-2">
                            <span className="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center text-white text-sm font-medium">
                              {card.assignee.full_name?.charAt(0).toUpperCase() || '?'}
                            </span>
                            <span className="text-sm font-medium">{card.assignee.full_name}</span>
                            {canEdit && (
                              <button
                                onClick={handleUnassign}
                                className="ml-auto text-xs text-gray-500 hover:text-red-500"
                              >
                                Remove
                              </button>
                            )}
                          </div>
                        </div>
                      )}

                      {card.reporter && (
                        <div>
                          <h4 className="text-xs font-semibold text-gray-500 uppercase mb-2">
                            Reporter
                          </h4>
                          <div className="flex items-center gap-2">
                            <span className="w-8 h-8 bg-green-500 rounded-full flex items-center justify-center text-white text-sm font-medium">
                              {card.reporter.full_name?.charAt(0).toUpperCase() || '?'}
                            </span>
                            <div>
                              <span className="text-sm font-medium">{card.reporter.full_name}</span>
                              <p className="text-xs text-gray-500">
                                Created {new Date(card.created_at).toLocaleDateString()}
                              </p>
                            </div>
                          </div>
                        </div>
                      )}

                      {card.due_date && (
                        <div>
                          <h4 className="text-xs font-semibold text-gray-500 uppercase mb-2">
                            Due date
                          </h4>
                          <div
                            onClick={canEdit ? handleToggleComplete : undefined}
                            className={`flex items-center gap-2 px-3 py-1.5 rounded text-sm ${
                              card.is_completed
                                ? 'bg-green-500 text-white'
                                : new Date(card.due_date) < new Date()
                                ? 'bg-red-500 text-white'
                                : 'bg-gray-200'
                            } ${canEdit ? 'cursor-pointer' : ''}`}
                          >
                            {canEdit && (
                              <input
                                type="checkbox"
                                checked={card.is_completed}
                                onChange={() => {}}
                                className="rounded"
                              />
                            )}
                            {new Date(card.due_date).toLocaleDateString('en-US', {
                              month: 'short',
                              day: 'numeric',
                              year: 'numeric',
                            })}
                            {card.is_completed && <span className="ml-1">complete</span>}
                          </div>
                        </div>
                      )}

                      <div>
                        <div className="flex items-center gap-2 mb-2">
                          <AlignLeft size={20} className="text-gray-500" />
                          <h4 className="font-semibold">Description</h4>
                        </div>
                        {isEditingDesc && canEdit ? (
                          <div>
                            <textarea
                              value={description}
                              onChange={(e) => setDescription(e.target.value)}
                              className="w-full p-3 border rounded-lg resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
                              rows={4}
                              placeholder="Add a more detailed description..."
                              autoFocus
                            />
                            <div className="flex gap-2 mt-2">
                              <button
                                onClick={handleDescriptionSave}
                                className="px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
                              >
                                Save
                              </button>
                              <button
                                onClick={() => {
                                  setDescription(card.description || '');
                                  setIsEditingDesc(false);
                                }}
                                className="px-3 py-1.5 text-gray-600 hover:bg-gray-200 rounded"
                              >
                                Cancel
                              </button>
                            </div>
                          </div>
                        ) : (
                          <div
                            onClick={() => canEdit && setIsEditingDesc(true)}
                            className={`p-3 bg-gray-200 rounded-lg min-h-[80px] ${canEdit ? 'cursor-pointer hover:bg-gray-300' : ''}`}
                          >
                            {description || (
                              <span className="text-gray-500">
                                {canEdit ? 'Add a more detailed description...' : 'No description'}
                              </span>
                            )}
                          </div>
                        )}
                      </div>
                    </div>

                    {canEdit && (
                      <div className="w-48 space-y-2">
                        <h4 className="text-xs font-semibold text-gray-500 uppercase mb-3">
                          Add to card
                        </h4>

                        <div className="relative">
                          <button
                            onClick={() => setShowMemberPicker(!showMemberPicker)}
                            className="flex items-center gap-2 w-full px-3 py-2 bg-gray-200 rounded hover:bg-gray-300 text-sm"
                          >
                            <User size={16} />
                            Members
                          </button>
                          {showMemberPicker && (
                            <>
                              <div
                                className="fixed inset-0 z-10"
                                onClick={() => setShowMemberPicker(false)}
                              />
                              <div className="absolute left-0 top-full mt-1 bg-white rounded-lg shadow-lg border p-2 z-20 w-64">
                                <h5 className="font-semibold text-sm mb-2 px-2">Board members</h5>
                                {members.map((member) => (
                                  <button
                                    key={member.user.id}
                                    onClick={() => handleAssign(member.user.id)}
                                    className="flex items-center gap-2 w-full px-2 py-1.5 hover:bg-gray-100 rounded text-sm"
                                  >
                                    <span className="w-6 h-6 bg-gray-300 rounded-full flex items-center justify-center text-xs font-medium">
                                      {member.user.full_name?.charAt(0).toUpperCase()}
                                    </span>
                                    {member.user.full_name}
                                    {card.assignee?.id === member.user.id && (
                                      <span className="ml-auto text-blue-600">✓</span>
                                    )}
                                  </button>
                                ))}
                                {card.assignee && (
                                  <button
                                    onClick={handleUnassign}
                                    className="w-full px-2 py-1.5 mt-1 text-sm text-red-600 hover:bg-red-50 rounded"
                                  >
                                    Remove assignee
                                  </button>
                                )}
                              </div>
                            </>
                          )}
                        </div>

                        <div className="relative">
                          <button
                            onClick={() => setShowLabelPicker(!showLabelPicker)}
                            className="flex items-center gap-2 w-full px-3 py-2 bg-gray-200 rounded hover:bg-gray-300 text-sm"
                          >
                            <Tag size={16} />
                            Labels
                          </button>
                          {showLabelPicker && (
                            <LabelPicker
                              boardLabels={boardLabels}
                              cardLabels={card.labels || []}
                              onToggle={handleLabelToggle}
                              onClose={() => setShowLabelPicker(false)}
                            />
                          )}
                        </div>

                        <div className="relative">
                          <button
                            onClick={() => setShowDatePicker(!showDatePicker)}
                            className="flex items-center gap-2 w-full px-3 py-2 bg-gray-200 rounded hover:bg-gray-300 text-sm"
                          >
                            <Calendar size={16} />
                            Dates
                          </button>
                          {showDatePicker && (
                            <>
                              <div
                                className="fixed inset-0 z-10"
                                onClick={() => setShowDatePicker(false)}
                              />
                              <div className="absolute left-0 top-full mt-1 bg-white rounded-lg shadow-lg border p-3 z-20">
                                <input
                                  type="datetime-local"
                                  value={
                                    card.due_date
                                      ? new Date(card.due_date).toISOString().slice(0, 16)
                                      : ''
                                  }
                                  onChange={handleDueDateChange}
                                  className="px-3 py-2 border rounded"
                                />
                              </div>
                            </>
                          )}
                        </div>

                        <div className="pt-4">
                          <h4 className="text-xs font-semibold text-gray-500 uppercase mb-3">
                            Actions
                          </h4>

                          <div className="space-y-1">
                            <select
                              value={card.priority}
                              onChange={(e) => handlePriorityChange(e.target.value)}
                              className="w-full px-3 py-2 bg-gray-200 rounded text-sm focus:outline-none"
                            >
                              {priorityOptions.map((opt) => (
                                <option key={opt.value} value={opt.value}>
                                  Priority: {opt.label}
                                </option>
                              ))}
                            </select>

                            <button
                              onClick={handleArchive}
                              className="flex items-center gap-2 w-full px-3 py-2 bg-gray-200 rounded hover:bg-red-100 hover:text-red-600 text-sm"
                            >
                              <Archive size={16} />
                              Archive
                            </button>
                          </div>
                        </div>
                      </div>
                    )}
                  </div>
                )}

                {activeTab === 'checklists' && <ChecklistSection cardId={card.id} canEdit={canEdit} />}

                {activeTab === 'attachments' && <AttachmentSection cardId={card.id} canEdit={canEdit} />}

                {activeTab === 'comments' && <CommentSection cardId={card.id} canEdit={canEdit} members={members} />}

                {activeTab === 'activity' && <ActivitySection cardId={card.id} />}
              </div>
            </>
          ) : (
            <div className="p-8 text-center text-gray-500">Card not found</div>
          )}
        </div>
      </div>
    </div>
  );
}
