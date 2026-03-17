import { BrowserRouter, Routes, Route } from 'react-router-dom'
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
import Dashboard from './pages/Dashboard'
import Settings from './pages/Settings'
import UserProfile from './pages/UserProfile'
import NotFound from './pages/NotFound'
import './styles/global.css'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/feed" element={<PublicFeed />} />
        <Route element={<AppLayout />}>
          <Route path="/policies/:slug" element={<PolicyDetail />} />
          <Route path="/policies/new" element={<CreatePolicy />} />
          <Route path="/search" element={<Search />} />
          <Route path="/bookmarks" element={<Bookmarks />} />
          <Route path="/representatives" element={<Representatives />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="/u/:username" element={<UserProfile />} />
          <Route path="*" element={<NotFound />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
