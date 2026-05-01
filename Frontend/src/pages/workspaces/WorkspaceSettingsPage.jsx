import { useParams } from 'react-router-dom';

export default function WorkspaceSettingsPage() {
  const { slug } = useParams();

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Workspace Settings</h1>
      <p className="text-gray-600">Settings for workspace: {slug}</p>
      <p className="text-gray-500 mt-4">Coming soon...</p>
    </div>
  );
}
