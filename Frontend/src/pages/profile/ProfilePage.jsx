import { useSelector } from 'react-redux';
import { User } from 'lucide-react';

export default function ProfilePage() {
  const { user } = useSelector((state) => state.auth);

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Profile</h1>

      <div className="card p-6">
        <div className="flex items-center gap-4 mb-6">
          <div className="w-20 h-20 rounded-full bg-gray-200 flex items-center justify-center overflow-hidden">
            {user?.avatar_url ? (
              <img src={user.avatar_url} alt={user.full_name} className="w-full h-full object-cover" />
            ) : (
              <User className="w-10 h-10 text-gray-400" />
            )}
          </div>
          <div>
            <h2 className="text-xl font-semibold text-gray-900">{user?.full_name}</h2>
            <p className="text-gray-500">{user?.email}</p>
          </div>
        </div>

        <p className="text-gray-500">Profile editing coming soon...</p>
      </div>
    </div>
  );
}
