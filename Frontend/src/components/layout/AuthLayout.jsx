import { Link } from 'react-router-dom';
import { Kanban } from 'lucide-react';

export default function AuthLayout({ children }) {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex flex-col">
      <header className="p-4">
        <Link to="/" className="flex items-center gap-2 text-trello-blue font-bold text-xl">
          <Kanban className="w-8 h-8" />
          <span>TaskFlow</span>
        </Link>
      </header>

      <main className="flex-1 flex items-center justify-center p-4">
        <div className="w-full max-w-md">
          <div className="bg-white rounded-lg shadow-lg p-8">{children}</div>
        </div>
      </main>

      <footer className="p-4 text-center text-gray-500 text-sm">
        <p>&copy; 2024 TaskFlow. All rights reserved.</p>
      </footer>
    </div>
  );
}
