import { useState, useRef, useEffect, useCallback } from 'react';

export default function MentionInput({
  value,
  onChange,
  onSubmit,
  members = [],
  placeholder = 'Write a comment...',
  className = '',
  disabled = false,
}) {
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [suggestions, setSuggestions] = useState([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [mentionSearch, setMentionSearch] = useState('');
  const [mentionStart, setMentionStart] = useState(-1);
  const inputRef = useRef(null);
  const suggestionsRef = useRef(null);

  const filterMembers = useCallback(
    (search) => {
      if (!search) return members;
      const lowerSearch = search.toLowerCase();
      return members.filter(
        (m) =>
          m.user?.full_name?.toLowerCase().includes(lowerSearch) ||
          m.user?.email?.toLowerCase().includes(lowerSearch)
      );
    },
    [members]
  );

  const handleInputChange = (e) => {
    const newValue = e.target.value;
    const cursorPos = e.target.selectionStart;

    onChange(newValue);

    // Check for @ mention trigger
    const textBeforeCursor = newValue.slice(0, cursorPos);
    const atIndex = textBeforeCursor.lastIndexOf('@');

    if (atIndex !== -1) {
      const textAfterAt = textBeforeCursor.slice(atIndex + 1);
      // Only show suggestions if @ is at start or after a space, and no space after @
      const charBeforeAt = atIndex > 0 ? textBeforeCursor[atIndex - 1] : ' ';
      if ((charBeforeAt === ' ' || charBeforeAt === '\n' || atIndex === 0) && !textAfterAt.includes(' ')) {
        setMentionStart(atIndex);
        setMentionSearch(textAfterAt);
        setSuggestions(filterMembers(textAfterAt));
        setShowSuggestions(true);
        setSelectedIndex(0);
        return;
      }
    }

    setShowSuggestions(false);
    setMentionStart(-1);
  };

  const insertMention = (member) => {
    if (mentionStart === -1) return;

    const beforeMention = value.slice(0, mentionStart);
    const afterMention = value.slice(mentionStart + mentionSearch.length + 1);
    const mentionText = `@[${member.user.full_name}](${member.user.id}) `;
    const newValue = beforeMention + mentionText + afterMention;

    onChange(newValue);
    setShowSuggestions(false);
    setMentionStart(-1);
    setMentionSearch('');

    // Focus back on input
    setTimeout(() => {
      if (inputRef.current) {
        const newCursorPos = beforeMention.length + mentionText.length;
        inputRef.current.focus();
        inputRef.current.setSelectionRange(newCursorPos, newCursorPos);
      }
    }, 0);
  };

  const handleKeyDown = (e) => {
    if (showSuggestions && suggestions.length > 0) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        setSelectedIndex((prev) => (prev + 1) % suggestions.length);
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        setSelectedIndex((prev) => (prev - 1 + suggestions.length) % suggestions.length);
      } else if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        insertMention(suggestions[selectedIndex]);
      } else if (e.key === 'Escape') {
        setShowSuggestions(false);
      }
    } else if (e.key === 'Enter' && !e.shiftKey && onSubmit) {
      e.preventDefault();
      onSubmit();
    }
  };

  useEffect(() => {
    if (suggestionsRef.current && selectedIndex >= 0) {
      const selectedEl = suggestionsRef.current.children[selectedIndex];
      if (selectedEl) {
        selectedEl.scrollIntoView({ block: 'nearest' });
      }
    }
  }, [selectedIndex]);

  // Close suggestions on click outside
  useEffect(() => {
    const handleClickOutside = (e) => {
      if (
        suggestionsRef.current &&
        !suggestionsRef.current.contains(e.target) &&
        inputRef.current &&
        !inputRef.current.contains(e.target)
      ) {
        setShowSuggestions(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Render text with highlighted mentions
  const renderDisplayText = (text) => {
    const mentionRegex = /@\[([^\]]+)\]\([^)]+\)/g;
    const parts = [];
    let lastIndex = 0;
    let match;

    while ((match = mentionRegex.exec(text)) !== null) {
      if (match.index > lastIndex) {
        parts.push(text.slice(lastIndex, match.index));
      }
      parts.push(
        <span key={match.index} className="text-blue-600 font-medium">
          @{match[1]}
        </span>
      );
      lastIndex = match.index + match[0].length;
    }

    if (lastIndex < text.length) {
      parts.push(text.slice(lastIndex));
    }

    return parts;
  };

  return (
    <div className="relative flex-1">
      <input
        ref={inputRef}
        type="text"
        value={value}
        onChange={handleInputChange}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        disabled={disabled}
        className={`w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 ${className}`}
      />

      {showSuggestions && suggestions.length > 0 && (
        <div
          ref={suggestionsRef}
          className="absolute left-0 right-0 bottom-full mb-1 bg-white border border-gray-200 rounded-lg shadow-lg max-h-48 overflow-y-auto z-50"
        >
          {suggestions.map((member, index) => (
            <button
              key={member.user.id}
              type="button"
              onClick={() => insertMention(member)}
              className={`w-full flex items-center gap-3 px-3 py-2 text-left hover:bg-gray-50 ${
                index === selectedIndex ? 'bg-blue-50' : ''
              }`}
            >
              <div className="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center text-white text-sm font-medium flex-shrink-0">
                {member.user.full_name?.charAt(0).toUpperCase() || '?'}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-gray-900 truncate">{member.user.full_name}</p>
                <p className="text-xs text-gray-500 truncate">{member.user.email}</p>
              </div>
              <span className="text-xs text-gray-400 capitalize">{member.role}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

// Helper to convert mention format for display
export function formatMentionsForDisplay(text) {
  if (!text) return '';
  return text.replace(/@\[([^\]]+)\]\([^)]+\)/g, '@$1');
}

// Helper to extract mentioned user IDs
export function extractMentionedUserIds(text) {
  const regex = /@\[([^\]]+)\]\(([^)]+)\)/g;
  const userIds = [];
  let match;
  while ((match = regex.exec(text)) !== null) {
    userIds.push(match[2]);
  }
  return userIds;
}
