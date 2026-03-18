import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTheme } from '../contexts/ThemeContext'
import { useAuth, getAccessToken } from '../hooks/useAuth'
import { Button, Input, Textarea, Modal, Toast } from '../components/ui'
import styles from './Settings.module.css'

interface ProfileForm {
  username: string
  email: string
  displayName: string
  bio: string
}

interface PasswordForm {
  currentPassword: string
  newPassword: string
  confirmPassword: string
}

interface NotificationSettings {
  emailUpdates: boolean
  debateReplies: boolean
  campaignUpdates: boolean
}

export default function Settings() {
  const { theme, setTheme } = useTheme()
  const { user, refreshUser, logout } = useAuth()
  const navigate = useNavigate()

  // Profile form state
  const [profile, setProfile] = useState<ProfileForm>({
    username: '',
    email: '',
    displayName: '',
    bio: '',
  })
  const [profileErrors, setProfileErrors] = useState<Partial<ProfileForm>>({})
  const [profileLoading, setProfileLoading] = useState(false)

  // Password form state
  const [password, setPassword] = useState<PasswordForm>({
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
  })
  const [passwordErrors, setPasswordErrors] = useState<Partial<PasswordForm>>({})
  const [passwordLoading, setPasswordLoading] = useState(false)

  // Notification settings
  const [notifications, setNotifications] = useState<NotificationSettings>({
    emailUpdates: true,
    debateReplies: true,
    campaignUpdates: true,
  })

  // Delete account modal
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [deletePassword, setDeletePassword] = useState('')
  const [deleteError, setDeleteError] = useState('')
  const [deleteLoading, setDeleteLoading] = useState(false)

  // Toast
  const [toast, setToast] = useState<{ message: string; variant: 'success' | 'error' } | null>(null)

  // Load user profile on mount
  useEffect(() => {
    async function loadProfile() {
      const token = getAccessToken()
      if (!token || !user) return

      try {
        const res = await fetch(`/api/users/${user.username}`, {
          headers: { Authorization: `Bearer ${token}` },
        })
        if (res.ok) {
          const data = await res.json()
          setProfile({
            username: data.username || '',
            email: data.email || '',
            displayName: data.displayName || '',
            bio: data.bio || '',
          })
        }
      } catch {
        // Non-critical
      }
    }
    loadProfile()
  }, [user])

  // Load notification settings from localStorage
  useEffect(() => {
    const saved = localStorage.getItem('ll-notifications')
    if (saved) {
      try {
        setNotifications(JSON.parse(saved))
      } catch {
        // Ignore invalid JSON
      }
    }
  }, [])

  const validateProfile = (): boolean => {
    const errors: Partial<ProfileForm> = {}
    if (profile.username.length < 3 || profile.username.length > 30) {
      errors.username = 'Username must be 3-30 characters'
    } else if (!/^[a-zA-Z0-9_]+$/.test(profile.username)) {
      errors.username = 'Username can only contain letters, numbers, and underscores'
    }
    if (profile.email && !profile.email.includes('@')) {
      errors.email = 'Please enter a valid email address'
    }
    setProfileErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleProfileSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!validateProfile()) return

    setProfileLoading(true)
    try {
      const token = getAccessToken()
      const res = await fetch('/api/users/me', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(profile),
      })

      if (res.ok) {
        await refreshUser()
        setToast({ message: 'Profile updated successfully', variant: 'success' })
      } else {
        const data = await res.json()
        setToast({ message: data.error || 'Failed to update profile', variant: 'error' })
      }
    } catch {
      setToast({ message: 'Failed to update profile', variant: 'error' })
    } finally {
      setProfileLoading(false)
    }
  }

  const validatePassword = (): boolean => {
    const errors: Partial<PasswordForm> = {}
    if (!password.currentPassword) {
      errors.currentPassword = 'Current password is required'
    }
    if (password.newPassword.length < 8) {
      errors.newPassword = 'Password must be at least 8 characters'
    }
    if (password.newPassword !== password.confirmPassword) {
      errors.confirmPassword = 'Passwords do not match'
    }
    setPasswordErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handlePasswordSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!validatePassword()) return

    setPasswordLoading(true)
    try {
      const token = getAccessToken()
      const res = await fetch('/api/users/me/password', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          currentPassword: password.currentPassword,
          newPassword: password.newPassword,
        }),
      })

      if (res.ok) {
        setPassword({ currentPassword: '', newPassword: '', confirmPassword: '' })
        setToast({ message: 'Password changed successfully', variant: 'success' })
      } else {
        const data = await res.json()
        setToast({ message: data.error || 'Failed to change password', variant: 'error' })
      }
    } catch {
      setToast({ message: 'Failed to change password', variant: 'error' })
    } finally {
      setPasswordLoading(false)
    }
  }

  const handleNotificationChange = (key: keyof NotificationSettings) => {
    const updated = { ...notifications, [key]: !notifications[key] }
    setNotifications(updated)
    localStorage.setItem('ll-notifications', JSON.stringify(updated))
  }

  const handleDeleteAccount = async () => {
    if (!deletePassword) {
      setDeleteError('Password is required')
      return
    }

    setDeleteLoading(true)
    setDeleteError('')
    try {
      const token = getAccessToken()
      const res = await fetch('/api/users/me', {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ password: deletePassword }),
      })

      if (res.ok) {
        logout()
        navigate('/')
      } else {
        const data = await res.json()
        setDeleteError(data.error || 'Failed to delete account')
      }
    } catch {
      setDeleteError('Failed to delete account')
    } finally {
      setDeleteLoading(false)
    }
  }

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Settings</h1>
      <p className={styles.subtitle}>Manage your account settings and preferences.</p>

      {/* Toast notification */}
      {toast && (
        <div className={styles.toastContainer}>
          <Toast
            variant={toast.variant}
            onClose={() => setToast(null)}
            autoDismiss={4000}
          >
            {toast.message}
          </Toast>
        </div>
      )}

      {/* Appearance Section */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Appearance</h2>
        <div className={styles.card}>
          <div className={styles.settingRow}>
            <div className={styles.settingInfo}>
              <span className={styles.settingLabel}>Theme</span>
              <span className={styles.settingDesc}>
                Choose how Lobster Lobby looks to you.
              </span>
            </div>
            <div className={styles.themeSwitch}>
              <button
                className={`${styles.themeOption} ${theme === 'light' ? styles.themeOptionActive : ''}`}
                onClick={() => setTheme('light')}
                aria-pressed={theme === 'light'}
              >
                Light
              </button>
              <button
                className={`${styles.themeOption} ${theme === 'dark' ? styles.themeOptionActive : ''}`}
                onClick={() => setTheme('dark')}
                aria-pressed={theme === 'dark'}
              >
                Dark
              </button>
            </div>
          </div>
        </div>
      </section>

      {/* Account Section */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Account</h2>
        <div className={styles.card}>
          <form onSubmit={handleProfileSubmit} className={styles.form}>
            <Input
              label="Username"
              value={profile.username}
              onChange={(e) => setProfile({ ...profile, username: e.target.value })}
              error={profileErrors.username}
            />
            <Input
              label="Email"
              type="email"
              value={profile.email}
              onChange={(e) => setProfile({ ...profile, email: e.target.value })}
              error={profileErrors.email}
            />
            <Input
              label="Display Name"
              value={profile.displayName}
              onChange={(e) => setProfile({ ...profile, displayName: e.target.value })}
              hint="This is how your name appears to others"
            />
            <Textarea
              label="Bio"
              value={profile.bio}
              onChange={(e) => setProfile({ ...profile, bio: e.target.value })}
              rows={3}
              hint="A short description about yourself"
            />
            <div className={styles.formActions}>
              <Button type="submit" disabled={profileLoading}>
                {profileLoading ? 'Saving...' : 'Save Changes'}
              </Button>
            </div>
          </form>
        </div>
      </section>

      {/* Password Section */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Password</h2>
        <div className={styles.card}>
          <form onSubmit={handlePasswordSubmit} className={styles.form}>
            <Input
              label="Current Password"
              type="password"
              value={password.currentPassword}
              onChange={(e) => setPassword({ ...password, currentPassword: e.target.value })}
              error={passwordErrors.currentPassword}
            />
            <Input
              label="New Password"
              type="password"
              value={password.newPassword}
              onChange={(e) => setPassword({ ...password, newPassword: e.target.value })}
              error={passwordErrors.newPassword}
              hint="Must be at least 8 characters"
            />
            <Input
              label="Confirm New Password"
              type="password"
              value={password.confirmPassword}
              onChange={(e) => setPassword({ ...password, confirmPassword: e.target.value })}
              error={passwordErrors.confirmPassword}
            />
            <div className={styles.formActions}>
              <Button type="submit" disabled={passwordLoading}>
                {passwordLoading ? 'Changing...' : 'Change Password'}
              </Button>
            </div>
          </form>
        </div>
      </section>

      {/* Notifications Section */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Notifications</h2>
        <div className={styles.card}>
          <div className={styles.settingRow}>
            <div className={styles.settingInfo}>
              <span className={styles.settingLabel}>Email Updates</span>
              <span className={styles.settingDesc}>
                Receive updates about new policies and features
              </span>
            </div>
            <button
              className={`${styles.toggle} ${notifications.emailUpdates ? styles.toggleActive : ''}`}
              onClick={() => handleNotificationChange('emailUpdates')}
              aria-pressed={notifications.emailUpdates}
            >
              <span className={styles.toggleKnob} />
            </button>
          </div>
          <div className={styles.divider} />
          <div className={styles.settingRow}>
            <div className={styles.settingInfo}>
              <span className={styles.settingLabel}>Debate Replies</span>
              <span className={styles.settingDesc}>
                Get notified when someone replies to your comments
              </span>
            </div>
            <button
              className={`${styles.toggle} ${notifications.debateReplies ? styles.toggleActive : ''}`}
              onClick={() => handleNotificationChange('debateReplies')}
              aria-pressed={notifications.debateReplies}
            >
              <span className={styles.toggleKnob} />
            </button>
          </div>
          <div className={styles.divider} />
          <div className={styles.settingRow}>
            <div className={styles.settingInfo}>
              <span className={styles.settingLabel}>Campaign Updates</span>
              <span className={styles.settingDesc}>
                Stay informed about campaigns you've supported
              </span>
            </div>
            <button
              className={`${styles.toggle} ${notifications.campaignUpdates ? styles.toggleActive : ''}`}
              onClick={() => handleNotificationChange('campaignUpdates')}
              aria-pressed={notifications.campaignUpdates}
            >
              <span className={styles.toggleKnob} />
            </button>
          </div>
        </div>
      </section>

      {/* Danger Zone */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitleDanger}>Danger Zone</h2>
        <div className={styles.cardDanger}>
          <div className={styles.settingRow}>
            <div className={styles.settingInfo}>
              <span className={styles.settingLabel}>Delete Account</span>
              <span className={styles.settingDesc}>
                Permanently delete your account and all associated data. This action cannot be undone.
              </span>
            </div>
            <Button variant="danger" onClick={() => setShowDeleteModal(true)}>
              Delete Account
            </Button>
          </div>
        </div>
      </section>

      {/* Delete Account Modal */}
      <Modal
        isOpen={showDeleteModal}
        onClose={() => {
          setShowDeleteModal(false)
          setDeletePassword('')
          setDeleteError('')
        }}
        title="Delete Account"
      >
        <div className={styles.modalContent}>
          <p className={styles.modalText}>
            Are you sure you want to delete your account? This action is permanent and cannot be undone.
            All your data, including your policies, comments, and bookmarks will be permanently removed.
          </p>
          <Input
            label="Enter your password to confirm"
            type="password"
            value={deletePassword}
            onChange={(e) => setDeletePassword(e.target.value)}
            error={deleteError}
          />
          <div className={styles.modalActions}>
            <Button
              variant="ghost"
              onClick={() => {
                setShowDeleteModal(false)
                setDeletePassword('')
                setDeleteError('')
              }}
            >
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={handleDeleteAccount}
              disabled={deleteLoading}
            >
              {deleteLoading ? 'Deleting...' : 'Delete My Account'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  )
}
