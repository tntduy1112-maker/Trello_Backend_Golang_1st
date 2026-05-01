import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import workspaceService from '../../services/workspace.service';

const initialState = {
  workspaces: [],
  currentWorkspace: null,
  members: [],
  isLoading: false,
  error: null,
};

export const fetchWorkspaces = createAsyncThunk(
  'workspace/fetchAll',
  async (_, { rejectWithValue }) => {
    try {
      const response = await workspaceService.getAll();
      return response.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to fetch workspaces' });
    }
  }
);

export const fetchWorkspace = createAsyncThunk(
  'workspace/fetchOne',
  async (slug, { rejectWithValue }) => {
    try {
      const response = await workspaceService.getBySlug(slug);
      return response.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to fetch workspace' });
    }
  }
);

export const createWorkspace = createAsyncThunk(
  'workspace/create',
  async (data, { rejectWithValue }) => {
    try {
      const response = await workspaceService.create(data);
      return response.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to create workspace' });
    }
  }
);

export const updateWorkspace = createAsyncThunk(
  'workspace/update',
  async ({ slug, data }, { rejectWithValue }) => {
    try {
      const response = await workspaceService.update(slug, data);
      return response.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to update workspace' });
    }
  }
);

export const deleteWorkspace = createAsyncThunk(
  'workspace/delete',
  async (slug, { rejectWithValue }) => {
    try {
      await workspaceService.delete(slug);
      return slug;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to delete workspace' });
    }
  }
);

export const fetchWorkspaceMembers = createAsyncThunk(
  'workspace/fetchMembers',
  async (slug, { rejectWithValue }) => {
    try {
      const response = await workspaceService.getMembers(slug);
      return response.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to fetch members' });
    }
  }
);

export const inviteMember = createAsyncThunk(
  'workspace/inviteMember',
  async ({ slug, email, role }, { rejectWithValue }) => {
    try {
      const response = await workspaceService.inviteMember(slug, email, role);
      return response.data;
    } catch (error) {
      return rejectWithValue(error.response?.data?.error || { message: 'Failed to invite member' });
    }
  }
);

const workspaceSlice = createSlice({
  name: 'workspace',
  initialState,
  reducers: {
    clearError: (state) => {
      state.error = null;
    },
    clearCurrentWorkspace: (state) => {
      state.currentWorkspace = null;
      state.members = [];
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchWorkspaces.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(fetchWorkspaces.fulfilled, (state, action) => {
        state.isLoading = false;
        state.workspaces = action.payload;
      })
      .addCase(fetchWorkspaces.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload;
      })
      .addCase(fetchWorkspace.fulfilled, (state, action) => {
        state.currentWorkspace = action.payload;
      })
      .addCase(createWorkspace.fulfilled, (state, action) => {
        state.workspaces.push(action.payload);
      })
      .addCase(updateWorkspace.fulfilled, (state, action) => {
        const index = state.workspaces.findIndex((w) => w.slug === action.payload.slug);
        if (index !== -1) {
          state.workspaces[index] = action.payload;
        }
        if (state.currentWorkspace?.slug === action.payload.slug) {
          state.currentWorkspace = action.payload;
        }
      })
      .addCase(deleteWorkspace.fulfilled, (state, action) => {
        state.workspaces = state.workspaces.filter((w) => w.slug !== action.payload);
      })
      .addCase(fetchWorkspaceMembers.fulfilled, (state, action) => {
        state.members = action.payload;
      });
  },
});

export const { clearError, clearCurrentWorkspace } = workspaceSlice.actions;
export default workspaceSlice.reducer;
