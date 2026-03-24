import { Link } from 'react-router-dom'
import styles from './legal.module.css'

const LAST_UPDATED = 'March 24, 2026'

export default function TermsOfService() {
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <div className={styles.mascot}>🦞</div>
        <h1 className={styles.title}>Terms of Service</h1>
        <p className={styles.subtitle}>Last updated: {LAST_UPDATED}</p>
      </header>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>1. Acceptance of Terms</h2>
        <p className={styles.text}>
          Welcome to Lobster Lobby ("we," "us," or "our"). By accessing or using our platform at
          lobsterlobby.com (the "Service"), you agree to be bound by these Terms of Service. If
          you do not agree to these terms, please do not use the Service.
        </p>
        <p className={styles.text}>
          These Terms apply to all visitors, users, and others who access or use the Service.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>2. Description of Service</h2>
        <p className={styles.text}>
          Lobster Lobby is a community platform for political discussion and civic engagement. The
          Service allows users to discuss policies, engage with representatives, participate in
          debates, run campaigns, and engage with other members of the community.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>3. User Accounts</h2>
        <p className={styles.text}>
          To use certain features of the Service, you must register for an account. You agree to:
        </p>
        <ul className={styles.list}>
          <li>Provide accurate, current, and complete information during registration</li>
          <li>Maintain and promptly update your account information</li>
          <li>Keep your password secure and confidential</li>
          <li>Accept responsibility for all activity that occurs under your account</li>
          <li>Notify us immediately of any unauthorized use of your account</li>
        </ul>
        <p className={styles.text}>
          You must be at least 13 years old to create an account. By creating an account, you
          represent that you meet this age requirement.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>4. Acceptable Use</h2>
        <p className={styles.text}>
          You agree not to use the Service to:
        </p>
        <ul className={styles.list}>
          <li>Post content that is unlawful, harmful, threatening, abusive, harassing, defamatory, or otherwise objectionable</li>
          <li>Impersonate any person, organization, or entity</li>
          <li>Spam, solicit, or post unsolicited commercial content</li>
          <li>Attempt to gain unauthorized access to the Service or other users' accounts</li>
          <li>Interfere with or disrupt the integrity or performance of the Service</li>
          <li>Engage in coordinated inauthentic behavior or manipulation of civic discourse</li>
          <li>Violate any applicable local, state, national, or international law</li>
        </ul>
        <p className={styles.text}>
          We reserve the right to remove any content and suspend or terminate any account that
          violates these terms.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>5. User Content</h2>
        <p className={styles.text}>
          You retain ownership of content you post on Lobster Lobby. By submitting content, you
          grant us a non-exclusive, royalty-free, worldwide license to use, display, and distribute
          your content in connection with operating the Service.
        </p>
        <p className={styles.text}>
          You represent that you have all rights necessary to grant this license and that your
          content does not violate any third-party rights or applicable laws.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>6. Moderation</h2>
        <p className={styles.text}>
          We reserve the right — but not the obligation — to monitor, edit, or remove any content
          that we determine, in our sole discretion, violates these Terms or is otherwise harmful
          to the community. We are not liable for any failure to remove, or any delay in removing,
          harmful content.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>7. Intellectual Property</h2>
        <p className={styles.text}>
          The Service and its original content (excluding user-submitted content), features, and
          functionality are and will remain the exclusive property of Lobster Lobby and its
          licensors. Our trademarks and trade dress may not be used in connection with any product
          or service without our prior written consent.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>8. Disclaimers &amp; Limitation of Liability</h2>
        <p className={styles.text}>
          The Service is provided on an "AS IS" and "AS AVAILABLE" basis without warranties of any
          kind, either express or implied. We do not warrant that the Service will be uninterrupted,
          error-free, or free of harmful components.
        </p>
        <p className={styles.text}>
          To the fullest extent permitted by law, Lobster Lobby shall not be liable for any
          indirect, incidental, special, consequential, or punitive damages arising from your use
          of, or inability to use, the Service.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>9. Governing Law</h2>
        <p className={styles.text}>
          These Terms of Service shall be governed by and construed in accordance with the laws of
          the State of Michigan, United States, without regard to its conflict of law provisions.
          Any dispute arising from or relating to these Terms or the Service shall be subject to
          the exclusive jurisdiction of the state and federal courts located in the State of
          Michigan, United States.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>10. Changes to Terms</h2>
        <p className={styles.text}>
          We reserve the right to modify these Terms at any time. We will notify users of material
          changes by updating the "Last updated" date at the top of this page. Your continued use
          of the Service after any changes constitutes your acceptance of the new Terms.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>11. Contact Us</h2>
        <p className={styles.text}>
          If you have any questions about these Terms, please contact us at:
        </p>
        <ul className={styles.list}>
          <li>Email: legal@lobsterlobby.com</li>
          <li>Platform: Lobster Lobby</li>
        </ul>
      </section>

      <footer className={styles.footer}>
        <p className={styles.footerText}>
          Also see our <Link to="/privacy">Privacy Policy</Link> for information about how we
          handle your data.
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
