import React from 'react';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { Home, Settings, Users, BarChart3, FileText } from 'lucide-react';

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

const Sidebar: React.FC<SidebarProps> = ({ isOpen, onClose }) => {
  const menuItems = [
    { icon: Home, label: 'Dashboard', href: '/dashboard' },
    { icon: Users, label: 'Users', href: '/users' },
    { icon: BarChart3, label: 'Analytics', href: '/analytics' },
    { icon: FileText, label: 'Reports', href: '/reports' },
    { icon: Settings, label: 'Settings', href: '/settings' },
  ];

  return (
    <Sheet open={isOpen} onOpenChange={onClose}>
      <SheetContent side="left" className="w-80 bg-gray-900 border-gray-800">
        <SheetHeader>
          <SheetTitle className="text-white">Navigation</SheetTitle>
        </SheetHeader>
        <div className="mt-6 space-y-2">
          {menuItems.map((item) => (
            <button
              key={item.label}
              className="w-full flex items-center space-x-3 px-4 py-3 text-gray-300 hover:text-white hover:bg-gray-800 rounded-lg transition-colors"
            >
              <item.icon className="w-5 h-5" />
              <span>{item.label}</span>
            </button>
          ))}
        </div>
      </SheetContent>
    </Sheet>
  );
};

export default Sidebar;