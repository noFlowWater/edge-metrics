import { NavLink, Outlet } from 'react-router';

export function Layout() {
  return (
    <div className="min-h-screen" style={{ backgroundColor: 'var(--color-pastel-bg)' }}>
      <nav style={{ backgroundColor: 'var(--color-pastel-card)', borderBottom: '1px solid var(--color-pastel-border)' }}>
        <div className="max-w-6xl mx-auto px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center gap-12">
              <span className="text-xl font-bold" style={{ color: 'var(--color-pastel-primary)' }}>
                Edge Metrics
              </span>
              <div className="flex items-center gap-2">
                <NavLink
                  to="/"
                  end
                  className={({ isActive }) =>
                    `px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                      isActive ? '' : ''
                    }`
                  }
                  style={({ isActive }) => ({
                    backgroundColor: isActive ? 'var(--color-pastel-primary)' : 'transparent',
                    color: isActive ? 'white' : 'var(--color-pastel-text-light)',
                  })}
                >
                  Dashboard
                </NavLink>
                <NavLink
                  to="/devices"
                  className={({ isActive }) =>
                    `px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                      isActive ? '' : 'hover:bg-opacity-10'
                    }`
                  }
                  style={({ isActive }) => ({
                    backgroundColor: isActive ? 'var(--color-pastel-primary)' : 'transparent',
                    color: isActive ? 'white' : 'var(--color-pastel-text-light)',
                  })}
                >
                  Devices
                </NavLink>
                <NavLink
                  to="/kubernetes"
                  className={({ isActive }) =>
                    `px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                      isActive ? '' : 'hover:bg-opacity-10'
                    }`
                  }
                  style={({ isActive }) => ({
                    backgroundColor: isActive ? 'var(--color-pastel-primary)' : 'transparent',
                    color: isActive ? 'white' : 'var(--color-pastel-text-light)',
                  })}
                >
                  Kubernetes
                </NavLink>
              </div>
            </div>
          </div>
        </div>
      </nav>
      <main className="max-w-6xl mx-auto py-10 px-6 lg:px-8">
        <Outlet />
      </main>
    </div>
  );
}
