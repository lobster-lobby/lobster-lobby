import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import { RequireAuth } from './components/RequireAuth'
import { RequireAdmin } from './components/RequireAdmin'
import { AppLayout } from './layouts/AppLayout'
import Home from './pages/Home'
import Login from './pages/Login'
import Register from './pages/Register'
import PublicFeed from './pages/PublicFeed'
import PolicyDetail from './pages/PolicyDetail'
import CreatePolicy from './pages/CreatePolicy'
import Search from './pages/Search'
import Bookmarks from './pages/Bookmarks'
import Representatives from './pages/Representatives'
import CampaignDetail from './pages/CampaignDetail'
import Dashboard from './pages/Dashboard'
import Settings from './pages/Settings'
import UserProfile from './pages/UserProfile'
import Debates from './pages/Debates'
import DebateDetail from './pages/DebateDetail'
import AdminModeration from './pages/AdminModeration'
import NotFound from './pages/NotFound'
import './styles/global.css'

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/feed" element={<PublicFeed />} />
          <Route element={<AppLayout />}>
            <Route path="/policies/:slug" element={<PolicyDetail />} />
            <Route
              path="/policies/new"
              element={
                <RequireAuth>
                  <CreatePolicy />
                </RequireAuth>
              }
            />
            <Route path="/search" element={<Search />} />
            <Route
              path="/bookmarks"
              element={
                <RequireAuth>
                  <Bookmarks />
                </RequireAuth>
              }
            />
            <Route path="/representatives" element={<Representatives />} />
            <Route path="/debates" element={<Debates />} />
            <Route path="/debates/:slug" element={<DebateDetail />} />
            <Route path="/campaigns/:slug" element={<CampaignDetail />} />
            <Route
              path="/dashboard"
              element={
                <RequireAuth>
                  <Dashboard />
                </RequireAuth>
              }
            />
            <Route
              path="/settings"
              element={
                <RequireAuth>
                  <Settings />
                </RequireAuth>
              }
            />
            <Route
              path="/admin/moderation"
              element={
                <RequireAdmin>
                  <AdminModeration />
                </RequireAdmin>
              }
            />
            <Route path="/u/:username" element={<UserProfile />} />
            <Route path="*" element={<NotFound />} />
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  )
}
