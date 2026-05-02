import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import boardService from '../../services/board.service';
import listService from '../../services/list.service';
import cardService from '../../services/card.service';

const initialState = {
  boards: [],
  currentBoard: null,
  lists: [],
  labels: [],
  members: [],
  openCardId: null,
  cardNavContext: null,
  isLoading: false,
  error: null,
};

export const fetchBoards = createAsyncThunk(
  'board/fetchAll',
  async (workspaceSlug, { rejectWithValue }) => {
    try {
      const response = await boardService.getByWorkspace(workspaceSlug);
      return response.data.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to fetch boards' });
    }
  }
);

export const fetchBoard = createAsyncThunk(
  'board/fetchOne',
  async (boardId, { rejectWithValue }) => {
    try {
      const response = await boardService.getById(boardId);
      return response.data.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to fetch board' });
    }
  }
);

export const createBoard = createAsyncThunk(
  'board/create',
  async ({ workspaceSlug, data }, { rejectWithValue }) => {
    try {
      const response = await boardService.create(workspaceSlug, data);
      return response.data.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to create board' });
    }
  }
);

export const updateBoard = createAsyncThunk(
  'board/update',
  async ({ boardId, data }, { rejectWithValue }) => {
    try {
      const response = await boardService.update(boardId, data);
      return response.data.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to update board' });
    }
  }
);

export const createList = createAsyncThunk(
  'board/createList',
  async ({ boardId, title }, { rejectWithValue }) => {
    try {
      const response = await listService.create(boardId, title);
      return response.data.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to create list' });
    }
  }
);

export const updateList = createAsyncThunk(
  'board/updateList',
  async ({ listId, data }, { rejectWithValue }) => {
    try {
      const response = await listService.update(listId, data);
      return response.data.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to update list' });
    }
  }
);

export const moveList = createAsyncThunk(
  'board/moveList',
  async ({ listId, position }, { rejectWithValue }) => {
    try {
      const response = await listService.move(listId, position);
      return response.data.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to move list' });
    }
  }
);

export const deleteList = createAsyncThunk(
  'board/deleteList',
  async (listId, { rejectWithValue }) => {
    try {
      await listService.archive(listId);
      return listId;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to delete list' });
    }
  }
);

export const createCard = createAsyncThunk(
  'board/createCard',
  async ({ listId, title }, { rejectWithValue }) => {
    try {
      const response = await cardService.create(listId, title);
      return { ...response.data.data, list_id: listId };
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to create card' });
    }
  }
);

export const updateCard = createAsyncThunk(
  'board/updateCard',
  async ({ cardId, data }, { rejectWithValue }) => {
    try {
      const response = await cardService.update(cardId, data);
      return response.data.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to update card' });
    }
  }
);

export const moveCard = createAsyncThunk(
  'board/moveCard',
  async ({ cardId, listId, position }, { rejectWithValue }) => {
    try {
      const response = await cardService.move(cardId, listId, position);
      return response.data.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to move card' });
    }
  }
);

export const deleteCard = createAsyncThunk(
  'board/deleteCard',
  async (cardId, { rejectWithValue }) => {
    try {
      await cardService.archive(cardId);
      return cardId;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to delete card' });
    }
  }
);

export const fetchMembers = createAsyncThunk(
  'board/fetchMembers',
  async (boardId, { rejectWithValue }) => {
    try {
      const response = await boardService.getMembers(boardId);
      return response.data.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to fetch members' });
    }
  }
);

export const inviteMember = createAsyncThunk(
  'board/inviteMember',
  async ({ boardId, email, role }, { rejectWithValue }) => {
    try {
      const response = await boardService.inviteMember(boardId, email, role);
      return response.data.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to invite member' });
    }
  }
);

const boardSlice = createSlice({
  name: 'board',
  initialState,
  reducers: {
    clearError: (state) => {
      state.error = null;
    },
    clearCurrentBoard: (state) => {
      state.currentBoard = null;
      state.lists = [];
      state.labels = [];
      state.members = [];
    },
    setOpenCardId: (state, action) => {
      if (typeof action.payload === 'object' && action.payload !== null) {
        state.openCardId = action.payload.cardId;
        state.cardNavContext = action.payload.context || null;
      } else {
        state.openCardId = action.payload;
        state.cardNavContext = null;
      }
    },
    clearCardNavContext: (state) => {
      state.cardNavContext = null;
    },
    optimisticMoveCard: (state, action) => {
      const { cardId, fromListId, toListId, newPosition } = action.payload;
      const fromList = state.lists.find((l) => l.id === fromListId);
      const toList = state.lists.find((l) => l.id === toListId);
      if (fromList && toList) {
        const cardIndex = fromList.cards.findIndex((c) => c.id === cardId);
        if (cardIndex !== -1) {
          const [card] = fromList.cards.splice(cardIndex, 1);
          card.position = newPosition;
          card.list_id = toListId;
          toList.cards.push(card);
          toList.cards.sort((a, b) => a.position - b.position);
        }
      }
    },
    optimisticMoveList: (state, action) => {
      const { listId, newPosition } = action.payload;
      const list = state.lists.find((l) => l.id === listId);
      if (list) {
        list.position = newPosition;
        state.lists.sort((a, b) => a.position - b.position);
      }
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchBoards.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(fetchBoards.fulfilled, (state, action) => {
        state.isLoading = false;
        state.boards = action.payload;
      })
      .addCase(fetchBoards.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload;
      })
      .addCase(fetchBoard.pending, (state) => {
        state.isLoading = true;
      })
      .addCase(fetchBoard.fulfilled, (state, action) => {
        state.isLoading = false;
        state.currentBoard = action.payload;
        state.lists = action.payload.lists || [];
        state.labels = action.payload.labels || [];
        state.members = action.payload.members || [];
      })
      .addCase(fetchBoard.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload;
      })
      .addCase(createBoard.fulfilled, (state, action) => {
        state.boards.push(action.payload);
      })
      .addCase(createList.fulfilled, (state, action) => {
        state.lists.push({ ...action.payload, cards: [] });
        state.lists.sort((a, b) => a.position - b.position);
      })
      .addCase(updateList.fulfilled, (state, action) => {
        const index = state.lists.findIndex((l) => l.id === action.payload.id);
        if (index !== -1) {
          state.lists[index] = { ...state.lists[index], ...action.payload };
        }
      })
      .addCase(deleteList.fulfilled, (state, action) => {
        state.lists = state.lists.filter((l) => l.id !== action.payload);
      })
      .addCase(createCard.fulfilled, (state, action) => {
        const list = state.lists.find((l) => l.id === action.payload.list_id);
        if (list) {
          if (!list.cards) list.cards = [];
          list.cards.push(action.payload);
          list.cards.sort((a, b) => a.position - b.position);
        }
      })
      .addCase(updateCard.fulfilled, (state, action) => {
        for (const list of state.lists) {
          const cardIndex = list.cards?.findIndex((c) => c.id === action.payload.id);
          if (cardIndex !== undefined && cardIndex !== -1) {
            list.cards[cardIndex] = { ...list.cards[cardIndex], ...action.payload };
            break;
          }
        }
      })
      .addCase(deleteCard.fulfilled, (state, action) => {
        for (const list of state.lists) {
          const cardIndex = list.cards?.findIndex((c) => c.id === action.payload);
          if (cardIndex !== undefined && cardIndex !== -1) {
            list.cards.splice(cardIndex, 1);
            break;
          }
        }
      })
      .addCase(fetchMembers.fulfilled, (state, action) => {
        state.members = action.payload || [];
      })
      .addCase(inviteMember.fulfilled, (state, action) => {
        if (action.payload) {
          state.members.push(action.payload);
        }
      });
  },
});

export const {
  clearError,
  clearCurrentBoard,
  setOpenCardId,
  clearCardNavContext,
  optimisticMoveCard,
  optimisticMoveList,
} = boardSlice.actions;
export default boardSlice.reducer;
