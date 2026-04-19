import { Navbar } from '../../components/Navbar';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="dark min-h-screen bg-background">
      <Navbar />
      {children}
    </div>
  );
}
