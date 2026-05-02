import { useState, useEffect, useRef } from 'react';
import { Paperclip, Upload, Trash2, Image, Download, Star } from 'lucide-react';
import attachmentService from '../../services/attachment.service';

const FILE_ICONS = {
  'image/': Image,
  'application/pdf': () => <span className="text-red-500 font-bold text-xs">PDF</span>,
  default: Paperclip,
};

const getFileIcon = (mimeType) => {
  if (mimeType?.startsWith('image/')) return FILE_ICONS['image/'];
  return FILE_ICONS[mimeType] || FILE_ICONS.default;
};

const formatFileSize = (bytes) => {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
};

export default function AttachmentSection({ cardId, onCoverChange, canEdit = true }) {
  const [attachments, setAttachments] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isUploading, setIsUploading] = useState(false);
  const fileInputRef = useRef(null);

  useEffect(() => {
    loadAttachments();
  }, [cardId]);

  const loadAttachments = async () => {
    try {
      const res = await attachmentService.list(cardId);
      setAttachments(res.data.data || []);
    } catch (error) {
      console.error('Failed to load attachments:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleFileSelect = async (e) => {
    const file = e.target.files[0];
    if (!file) return;

    const maxSize = 10 * 1024 * 1024;
    if (file.size > maxSize) {
      alert('File too large. Maximum size is 10MB.');
      return;
    }

    setIsUploading(true);
    try {
      const res = await attachmentService.upload(cardId, file);
      setAttachments([res.data.data, ...attachments]);
    } catch (error) {
      console.error('Failed to upload file:', error);
      alert(error.response?.data?.error?.message || 'Failed to upload file');
    } finally {
      setIsUploading(false);
      if (fileInputRef.current) fileInputRef.current.value = '';
    }
  };

  const handleDelete = async (attachmentId) => {
    if (!confirm('Delete this attachment?')) return;

    try {
      await attachmentService.delete(attachmentId);
      setAttachments(attachments.filter((a) => a.id !== attachmentId));
    } catch (error) {
      console.error('Failed to delete attachment:', error);
    }
  };

  const handleSetCover = async (attachmentId) => {
    try {
      await attachmentService.setCover(attachmentId);
      setAttachments(
        attachments.map((a) => ({
          ...a,
          is_cover: a.id === attachmentId,
        }))
      );
      if (onCoverChange) onCoverChange();
    } catch (error) {
      console.error('Failed to set cover:', error);
      alert(error.response?.data?.error?.message || 'Failed to set cover');
    }
  };

  const handleRemoveCover = async () => {
    try {
      await attachmentService.removeCover(cardId);
      setAttachments(attachments.map((a) => ({ ...a, is_cover: false })));
      if (onCoverChange) onCoverChange();
    } catch (error) {
      console.error('Failed to remove cover:', error);
    }
  };

  const handleDownload = async (attachment) => {
    try {
      const res = await attachmentService.getDownloadUrl(attachment.id);
      window.open(res.data.data.url, '_blank');
    } catch (error) {
      console.error('Failed to get download URL:', error);
    }
  };

  const isImage = (mimeType) => mimeType?.startsWith('image/');

  const formatDate = (dateStr) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  return (
    <div>
      <div className="flex items-center gap-2 mb-3">
        <Paperclip size={20} className="text-gray-500" />
        <h4 className="font-semibold">Attachments</h4>
      </div>

      {canEdit && (
        <>
          <input
            ref={fileInputRef}
            type="file"
            onChange={handleFileSelect}
            className="hidden"
            accept="image/*,.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.zip,.rar"
          />

          <button
            onClick={() => fileInputRef.current?.click()}
            disabled={isUploading}
            className="flex items-center gap-2 w-full px-3 py-2 bg-gray-200 rounded hover:bg-gray-300 text-sm mb-3 disabled:opacity-50"
          >
            <Upload size={16} />
            {isUploading ? 'Uploading...' : 'Add attachment'}
          </button>
        </>
      )}

      {isLoading ? (
        <div className="text-center py-4 text-gray-500 text-sm">Loading attachments...</div>
      ) : attachments.length === 0 ? (
        <div className="text-center py-4 text-gray-500 text-sm">No attachments</div>
      ) : (
        <div className="space-y-2">
          {attachments.map((attachment) => (
            <div
              key={attachment.id}
              className="flex gap-3 p-2 bg-white rounded-lg shadow-sm group hover:bg-gray-50"
            >
              <div className="w-20 h-14 bg-gray-100 rounded flex items-center justify-center overflow-hidden flex-shrink-0">
                {isImage(attachment.mime_type) ? (
                  <img
                    src={attachment.url}
                    alt={attachment.original_name}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  (() => {
                    const IconComponent = getFileIcon(attachment.mime_type);
                    return typeof IconComponent === 'function' ? (
                      <IconComponent size={24} className="text-gray-400" />
                    ) : (
                      IconComponent
                    );
                  })()
                )}
              </div>

              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-medium text-sm truncate">{attachment.original_name}</span>
                  {attachment.is_cover && (
                    <span className="px-1.5 py-0.5 bg-yellow-100 text-yellow-800 text-xs rounded">
                      Cover
                    </span>
                  )}
                </div>
                <div className="text-xs text-gray-500">
                  {formatFileSize(attachment.file_size)} - {formatDate(attachment.created_at)}
                </div>
                <div className="flex items-center gap-2 mt-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button
                    onClick={() => handleDownload(attachment)}
                    className="text-xs text-blue-600 hover:underline flex items-center gap-1"
                  >
                    <Download size={12} /> Download
                  </button>
                  {canEdit && isImage(attachment.mime_type) && (
                    <>
                      {attachment.is_cover ? (
                        <button
                          onClick={handleRemoveCover}
                          className="text-xs text-gray-600 hover:underline flex items-center gap-1"
                        >
                          <Star size={12} /> Remove cover
                        </button>
                      ) : (
                        <button
                          onClick={() => handleSetCover(attachment.id)}
                          className="text-xs text-gray-600 hover:underline flex items-center gap-1"
                        >
                          <Star size={12} /> Make cover
                        </button>
                      )}
                    </>
                  )}
                  {canEdit && (
                    <button
                      onClick={() => handleDelete(attachment.id)}
                      className="text-xs text-red-600 hover:underline flex items-center gap-1"
                    >
                      <Trash2 size={12} /> Delete
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
