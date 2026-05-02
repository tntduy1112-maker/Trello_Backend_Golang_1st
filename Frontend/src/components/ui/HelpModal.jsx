import { useState } from 'react';
import {
  X,
  HelpCircle,
  Layout,
  List,
  CreditCard,
  Tag,
  MessageSquare,
  CheckSquare,
  Paperclip,
  Bell,
  Users,
  Keyboard,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';

const helpSections = [
  {
    id: 'getting-started',
    title: 'Getting Started',
    icon: Layout,
    content: [
      {
        title: 'Create a Workspace',
        description: 'Click "Create Workspace" to create a new workspace where all your boards will live.',
      },
      {
        title: 'Create a Board',
        description: 'Inside your workspace, click "Create Board", give it a name, and choose a background color.',
      },
      {
        title: 'Invite Team Members',
        description: 'Click the "Invite" button on your board to add team members via email.',
      },
    ],
  },
  {
    id: 'lists',
    title: 'Working with Lists',
    icon: List,
    content: [
      {
        title: 'Add a List',
        description: 'Click "+ Add another list" on the right side of your board and enter a name.',
      },
      {
        title: 'Rename a List',
        description: 'Click on the list title to edit it inline.',
      },
      {
        title: 'Reorder Lists',
        description: 'Drag and drop lists to change their order on the board.',
      },
      {
        title: 'Archive a List',
        description: 'Click the menu icon on a list and select "Archive" to remove it.',
      },
    ],
  },
  {
    id: 'cards',
    title: 'Working with Cards',
    icon: CreditCard,
    content: [
      {
        title: 'Create a Card',
        description: 'Click "+ Add a card" at the bottom of any list and enter a title.',
      },
      {
        title: 'Edit a Card',
        description: 'Click on any card to open the detail modal where you can add descriptions, due dates, and more.',
      },
      {
        title: 'Move Cards',
        description: 'Drag and drop cards between lists or within a list to reorder them.',
      },
      {
        title: 'Set Due Dates',
        description: 'Open a card and click "Dates" to set a deadline. Overdue cards show in red.',
      },
      {
        title: 'Assign Members',
        description: 'Click "Members" in the card detail to assign the card to a team member.',
      },
    ],
  },
  {
    id: 'labels',
    title: 'Labels',
    icon: Tag,
    content: [
      {
        title: 'Add Labels',
        description: 'Open a card and click "Labels" to create or assign colored labels for categorization.',
      },
      {
        title: 'Create Custom Labels',
        description: 'Choose a color and optionally add a name to create meaningful labels.',
      },
      {
        title: 'Filter by Labels',
        description: 'Labels appear on cards in the board view for easy visual filtering.',
      },
    ],
  },
  {
    id: 'comments',
    title: 'Comments & Mentions',
    icon: MessageSquare,
    content: [
      {
        title: 'Add Comments',
        description: 'Open a card, go to the Comments tab, and write your comment.',
      },
      {
        title: '@Mention Team Members',
        description: 'Type @ followed by a name to mention someone. They will receive a notification.',
      },
      {
        title: 'Reply to Comments',
        description: 'Click "Reply" on any comment to create a threaded conversation.',
      },
      {
        title: 'Edit or Delete',
        description: 'You can edit or delete your own comments using the icons that appear on hover.',
      },
    ],
  },
  {
    id: 'checklists',
    title: 'Checklists',
    icon: CheckSquare,
    content: [
      {
        title: 'Create a Checklist',
        description: 'Open a card, go to the Checklists tab, and click "Add Checklist".',
      },
      {
        title: 'Add Items',
        description: 'Type item names and press Enter to add them to your checklist.',
      },
      {
        title: 'Track Progress',
        description: 'Check off items as you complete them. A progress bar shows overall completion.',
      },
    ],
  },
  {
    id: 'attachments',
    title: 'Attachments',
    icon: Paperclip,
    content: [
      {
        title: 'Upload Files',
        description: 'Open a card, go to the Attachments tab, and drag files or click to upload.',
      },
      {
        title: 'Preview & Download',
        description: 'Click on attachments to preview or download them.',
      },
      {
        title: 'Supported Files',
        description: 'Upload images, documents, PDFs, and other common file types.',
      },
    ],
  },
  {
    id: 'notifications',
    title: 'Notifications',
    icon: Bell,
    content: [
      {
        title: 'Notification Bell',
        description: 'Click the bell icon in the header to see all your notifications.',
      },
      {
        title: 'Real-time Updates',
        description: 'Receive instant notifications when someone mentions you or updates your cards.',
      },
      {
        title: 'Click to Navigate',
        description: 'Click on any notification to go directly to the relevant card and comment.',
      },
      {
        title: 'Mark as Read',
        description: 'Mark individual notifications or all as read using the buttons provided.',
      },
    ],
  },
  {
    id: 'collaboration',
    title: 'Team Collaboration',
    icon: Users,
    content: [
      {
        title: 'Member Roles',
        description: 'Owner: Full control | Admin: Manage members | Member: Edit content | Viewer: Read only',
      },
      {
        title: 'Invite by Email',
        description: 'Send email invitations to add new members to your board.',
      },
      {
        title: 'Activity Feed',
        description: 'View the Activity tab on any card to see who did what and when.',
      },
    ],
  },
  {
    id: 'shortcuts',
    title: 'Keyboard Shortcuts',
    icon: Keyboard,
    content: [
      {
        title: 'Enter',
        description: 'Confirm or submit forms and inputs.',
      },
      {
        title: 'Escape',
        description: 'Cancel or close modals and dropdowns.',
      },
      {
        title: '@ (in comments)',
        description: 'Open the mention picker to tag team members.',
      },
      {
        title: 'Arrow Keys',
        description: 'Navigate through dropdown menus and suggestions.',
      },
    ],
  },
];

export default function HelpModal({ onClose }) {
  const [expandedSection, setExpandedSection] = useState('getting-started');

  const toggleSection = (sectionId) => {
    setExpandedSection(expandedSection === sectionId ? null : sectionId);
  };

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="fixed inset-0 bg-black/50" onClick={onClose} />

        <div className="relative bg-white rounded-xl shadow-xl w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col">
          <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
            <div className="flex items-center gap-2">
              <HelpCircle className="w-6 h-6 text-blue-600" />
              <h2 className="text-xl font-semibold text-gray-900">Help & User Guide</h2>
            </div>
            <button
              onClick={onClose}
              className="p-1 rounded-full hover:bg-gray-100 transition-colors"
            >
              <X className="w-5 h-5 text-gray-500" />
            </button>
          </div>

          <div className="flex-1 overflow-y-auto p-4">
            <div className="space-y-2">
              {helpSections.map((section) => {
                const Icon = section.icon;
                const isExpanded = expandedSection === section.id;

                return (
                  <div key={section.id} className="border border-gray-200 rounded-lg overflow-hidden">
                    <button
                      onClick={() => toggleSection(section.id)}
                      className="w-full flex items-center justify-between px-4 py-3 bg-gray-50 hover:bg-gray-100 transition-colors"
                    >
                      <div className="flex items-center gap-3">
                        <Icon className="w-5 h-5 text-blue-600" />
                        <span className="font-medium text-gray-900">{section.title}</span>
                      </div>
                      {isExpanded ? (
                        <ChevronDown className="w-5 h-5 text-gray-400" />
                      ) : (
                        <ChevronRight className="w-5 h-5 text-gray-400" />
                      )}
                    </button>

                    {isExpanded && (
                      <div className="px-4 py-3 bg-white space-y-3">
                        {section.content.map((item, index) => (
                          <div key={index} className="pl-8">
                            <h4 className="font-medium text-gray-800 text-sm">{item.title}</h4>
                            <p className="text-gray-600 text-sm mt-0.5">{item.description}</p>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>

            <div className="mt-6 p-4 bg-blue-50 rounded-lg">
              <h3 className="font-semibold text-blue-900 mb-2">Need More Help?</h3>
              <p className="text-blue-800 text-sm">
                If you encounter any issues or have questions, please contact your workspace
                administrator or report bugs on our{' '}
                <a
                  href="https://github.com/tntduy1112-maker/Trello_Backend_Golang_1st/issues"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="underline hover:text-blue-600"
                >
                  GitHub Issues
                </a>{' '}
                page.
              </p>
            </div>
          </div>

          <div className="px-6 py-3 border-t border-gray-200 bg-gray-50 text-center">
            <p className="text-xs text-gray-500">TaskFlow v1.0 | May 2026</p>
          </div>
        </div>
      </div>
    </div>
  );
}
