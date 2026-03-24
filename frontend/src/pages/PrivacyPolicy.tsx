import { Link } from 'react-router-dom'
import styles from './legal.module.css'

const LAST_UPDATED = 'March 24, 2026'

export default function PrivacyPolicy() {
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <div className={styles.mascot}>🔒</div>
        <h1 className={styles.title}>Privacy Policy</h1>
        <p className={styles.subtitle}>Last updated: {LAST_UPDATED}</p>
      </header>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>1. Introduction</h2>
        <p className={styles.text}>
          Lobster Lobby ("we," "us," or "our") is committed to protecting your privacy. This
          Privacy Policy explains how we collect, use, disclose, and safeguard your information
          when you use our platform at lobsterlobby.com (the "Service").
        </p>
        <p className={styles.text}>
          Please read this policy carefully. By using the Service, you agree to the practices
          described here.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>2. Information We Collect</h2>
        <p className={styles.text}>We collect the following types of information:</p>
        <ul className={styles.list}>
          <li>
            <strong>Account information:</strong> Username, email address, and password (hashed)
            when you register
          </li>
          <li>
            <strong>Profile information:</strong> Any optional profile details you choose to provide
            (bio, location, avatar)
          </li>
          <li>
            <strong>User content:</strong> Posts, comments, votes, debate contributions, and
            campaign materials you create on the platform
          </li>
          <li>
            <strong>Usage data:</strong> Pages visited, features used, timestamps, and interactions
            within the Service
          </li>
          <li>
            <strong>Device and log data:</strong> IP address, browser type, operating system, and
            referring URLs
          </li>
        </ul>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>3. How We Use Your Information</h2>
        <p className={styles.text}>We use the information we collect to:</p>
        <ul className={styles.list}>
          <li>Provide, operate, and maintain the Service</li>
          <li>Create and manage your account</li>
          <li>Enable you to participate in discussions, debates, and campaigns</li>
          <li>Send you service-related notifications (account confirmations, moderation notices)</li>
          <li>Monitor and analyze usage to improve the Service</li>
          <li>Detect, prevent, and address technical issues and abuse</li>
          <li>Comply with legal obligations</li>
        </ul>
        <p className={styles.text}>
          We do not sell your personal information to third parties or use it for targeted
          advertising.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>4. Information Sharing</h2>
        <p className={styles.text}>
          We may share your information in the following limited circumstances:
        </p>
        <ul className={styles.list}>
          <li>
            <strong>Publicly posted content:</strong> Content you post publicly (discussions,
            debates, campaign positions) is visible to all users and may be indexed by search
            engines
          </li>
          <li>
            <strong>Service providers:</strong> Trusted third-party vendors who help us operate
            the Service (hosting, email delivery) and are bound by confidentiality agreements
          </li>
          <li>
            <strong>Legal requirements:</strong> When required by law, court order, or government
            authority
          </li>
          <li>
            <strong>Safety:</strong> To protect the rights, property, or safety of Lobster Lobby,
            our users, or the public
          </li>
          <li>
            <strong>Business transfers:</strong> In connection with a merger, acquisition, or sale
            of assets, with prior notice to affected users
          </li>
        </ul>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>5. Cookies &amp; Tracking</h2>
        <p className={styles.text}>
          We use cookies and similar tracking technologies to operate the Service:
        </p>
        <ul className={styles.list}>
          <li>
            <strong>Essential cookies:</strong> Required for authentication and session management.
            Without these, the Service cannot function.
          </li>
          <li>
            <strong>Preference cookies:</strong> Remember your settings such as theme (light/dark
            mode) and display preferences.
          </li>
          <li>
            <strong>Analytics cookies:</strong> Help us understand how users interact with the
            Service so we can improve it. Data is aggregated and anonymized where possible.
          </li>
        </ul>
        <p className={styles.text}>
          You can control cookies through your browser settings. Disabling essential cookies may
          prevent you from using parts of the Service.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>6. Data Retention</h2>
        <p className={styles.text}>
          We retain your personal information for as long as your account is active or as needed
          to provide the Service. You may request deletion of your account and associated data at
          any time (see Your Rights below). We may retain certain data for legal, safety, or
          anti-abuse purposes even after deletion.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>7. Data Security</h2>
        <p className={styles.text}>
          We implement industry-standard security measures including encryption in transit (HTTPS),
          hashed passwords, and access controls to protect your information. However, no method of
          transmission over the internet is 100% secure, and we cannot guarantee absolute security.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>8. Your Rights</h2>
        <p className={styles.text}>
          Depending on your location, you may have the following rights regarding your personal
          data:
        </p>
        <ul className={styles.list}>
          <li>
            <strong>Access:</strong> Request a copy of the personal data we hold about you
          </li>
          <li>
            <strong>Correction:</strong> Request correction of inaccurate or incomplete data
          </li>
          <li>
            <strong>Deletion:</strong> Request deletion of your account and personal data
          </li>
          <li>
            <strong>Portability:</strong> Request your data in a portable, machine-readable format
          </li>
          <li>
            <strong>Objection:</strong> Object to certain processing activities
          </li>
        </ul>
        <p className={styles.text}>
          To exercise these rights, contact us at privacy@lobsterlobby.com. We will respond within
          30 days.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>9. Children's Privacy</h2>
        <p className={styles.text}>
          The Service is not directed to children under 13. We do not knowingly collect personal
          information from children under 13. If we become aware that we have inadvertently
          collected such information, we will delete it promptly.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>10. Changes to This Policy</h2>
        <p className={styles.text}>
          We may update this Privacy Policy from time to time. We will notify you of material
          changes by updating the "Last updated" date and, where appropriate, by sending an email
          notification. Your continued use of the Service after changes are posted constitutes
          acceptance of the updated policy.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>11. Contact Us</h2>
        <p className={styles.text}>
          If you have questions or concerns about this Privacy Policy or our data practices,
          please contact us:
        </p>
        <ul className={styles.list}>
          <li>Email: privacy@lobsterlobby.com</li>
          <li>Platform: Lobster Lobby</li>
        </ul>
      </section>

      <footer className={styles.footer}>
        <p className={styles.footerText}>
          Also see our <Link to="/terms">Terms of Service</Link> for the rules governing use of
          the platform.
        </p>
        <div className={styles.navLinks}>
          <Link to="/" className={styles.backLink}>
            ← Home
          </Link>
          <Link to="/register" className={styles.backLink}>
            ← Back to Registration
          </Link>
        </div>
      </footer>
    </div>
  )
}
