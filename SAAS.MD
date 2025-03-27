# 📄 **CarFlow Hub — Product Requirements Document (PRD)**

## 🧠 Product Vision

> Build a lightweight, no-friction SaaS for small and mid-sized **auto repair shops in the US** who still rely on spreadsheets, whiteboards, or outdated systems.

**Goal:**  
Replace chaos with clarity by offering a simple tool to **track vehicles, schedule services, and keep customers coming back** — without overwhelming the business owner or the solo founder maintaining it.

## 🎯 Target Audience

- 🇺🇸 **Small auto repair shop owners** (1–10 staff)
- Independent shops struggling with administrative overhead
- Technicians spending too much time on paperwork instead of repairs
- Owners looking for affordable solutions ($20-50/month range)
- Businesses losing customers due to inconsistent follow-up
- Shops with 50-300 regular customers in their database

## 💡 Core Features (MVP Scope)

| Feature | Description | Priority |
|---------|-------------|----------|
| 🔐 **Authentication** | Supabase Auth with email/password login and optional Google OAuth | High |
| 🧑‍🔧 **User Isolation** | Multi-tenant architecture with complete data separation per business | High |
| 🚗 **Vehicle Inventory** | Register cars by plate, brand, model, year, color, VIN, and owner details | High |
| 👥 **Customer Management** | Store customer contact info, service history, and preferences | High |
| 📆 **Service Scheduling** | Calendar view with drag-and-drop appointment setting and status tracking | High |
| 🕒 **Upcoming Services** | Dashboard widget showing pending appointments with priority indicators | High |
| 📝 **Service History** | Complete vehicle service logs with parts, labor, and technician notes | Medium |
| 📬 **Email Reminders** | Automatic service reminders via Brevo with customizable templates | Medium |
| 💬 **SMS Notifications** | Optional text alerts for appointment confirmations and status updates (Pro plan) | Low |
| 💳 **Billing System** | Stripe subscription with seamless upgrades, downgrades, and cancelation | High |
| 📊 **Business Dashboard** | KPIs showing vehicle count, appointment load, revenue trends, and customer retention | Medium |
| 🔍 **Search & Filtering** | Quick lookup by plate number, customer name, or service type | Medium |
| ⚙️ **Self-service Portal** | Comprehensive documentation, video tutorials, and FAQ section | Medium |
| 🌐 **Landing Page** | Conversion-optimized marketing site with benefits, pricing, and testimonials | High |
| 📱 **Responsive Design** | Mobile-friendly interface that works on tablets and smartphones | High |

## 💰 Pricing Strategy

| Plan | Price | Features | Limits |
|------|-------|----------|--------|
| **Starter** | $29/mo | • All core features<br>• Email reminders<br>• Basic dashboard<br>• Standard support | • 1 user account<br>• Up to 50 vehicles<br>• 100 email reminders/mo |
| **Pro** | $49/mo | • Everything in Starter<br>• SMS notifications<br>• Advanced reporting<br>• Customizable templates<br>• Priority support | • Up to 3 user accounts<br>• Up to 200 vehicles<br>• 500 email reminders/mo<br>• 100 SMS/mo |
| **Enterprise** | $99/mo | • Everything in Pro<br>• White-labeling<br>• API access<br>• Custom integrations<br>• Dedicated account manager | • Up to 10 user accounts<br>• Unlimited vehicles<br>• 2,000 email reminders/mo<br>• 500 SMS/mo |

**Annual Discount:** 20% off for yearly subscriptions

## 🧱 Tech Stack

| Layer | Technology | Justification |
|-------|------------|---------------|
| **Frontend** | React SPA with Tailwind CSS | Fast development, mobile-responsive, lightweight |
| **UI Framework** | Lovable.dev components | Pre-built components save development time |
| **Deployment** | Vercel (frontend), Cloud Run (API) | Scalable, reliable, low maintenance |
| **Backend** | Golang API with Chi router | High performance, low resource usage |
| **Database** | PostgreSQL via Supabase | Robust relational DB with excellent tooling |
| **Authentication** | Supabase Auth (Pro Plan) | Secure, pre-built auth flows with minimal setup |
| **Payments** | Stripe with webhook integration | Industry standard payment processing |
| **Email** | Brevo (free 300/day tier) | Reliable email delivery with analytics |
| **SMS** | Twilio (Pro/Enterprise plans) | Industry standard SMS delivery |
| **Infrastructure** | Terraform with modular configuration | Infrastructure-as-code for predictable deployments |
| **CI/CD** | GitHub Actions with automated testing | Automated build/test/deploy pipeline |
| **Monitoring** | Google Cloud Monitoring + Sentry | Error tracking and performance monitoring |
| **Analytics** | Simple Analytics (privacy-focused) | Understanding user behavior without cookies |

## 🛡️ Non-Goals / What's Out (for now)

- ❌ **Native mobile app** — Focusing on responsive web experience first
- ❌ **AI/ML features** — Avoiding complexity and potential privacy concerns
- ❌ **Advanced analytics** — Basic reporting is sufficient for initial market fit
- ❌ **Chatbots or live chat** — Maintaining low-touch support model
- ❌ **Multi-language support** — US market focus (English only)
- ❌ **On-premise deployment** — Cloud-only to simplify maintenance
- ❌ **Inventory management** — Focusing on scheduling, not parts tracking
- ❌ **Accounting integration** — No QuickBooks/Xero connections initially
- ❌ **Online booking portal** — Starting with shop-side management only

## 🚀 Launch Plan

| Phase | Timeline | Key Deliverables | Success Metrics |
|-------|----------|------------------|-----------------|
| **Research & Design** | Days 1-5 | • User interviews<br>• Competitor analysis<br>• UI/UX wireframes<br>• Data schema design | • 5+ user interviews completed<br>• Design system established |
| **MVP Development** | Days 6-20 | • Authentication system<br>• Vehicle & customer database<br>• Basic scheduling<br>• Simple dashboard<br>• Stripe integration | • Core functionality working<br>• Test accounts created<br>• Payment processing verified |
| **Beta Testing** | Days 21-30 | • Onboarding 5-10 beta users<br>• Collecting feedback<br>• Fixing critical issues<br>• Implementing high-priority requests | • 5+ active beta users<br>• Initial feedback collected<br>• Major bugs resolved |
| **Public Launch** | Days 31-40 | • Marketing site live<br>• Documentation completed<br>• Email sequences set up<br>• Initial SEO implementation | • First 5 paying customers<br>• < 5% error rate in production |
| **Growth Phase** | Days 41-90 | • Targeted outreach to repair shops<br>• Content marketing (blog, guides)<br>• Referral program setup<br>• Customer success stories | • 50 paying customers<br>• $1,000+ MRR |

## 📊 KPIs & Success Metrics

| Metric | 30-Day Target | 90-Day Target | 180-Day Target |
|--------|---------------|---------------|----------------|
| **MRR** | $150 | $1,000 | $3,000 |
| **Paying Users** | 5 | 50 | 120 |
| **User Retention** | 90% | 92% | 95% |
| **Monthly Churn** | < 10% | < 8% | < 5% |
| **Infrastructure Cost** | < $5/mo | < $10/mo | < $25/mo |
| **Support Tickets** | < 5/week | < 10/week | < 20/week |
| **Feature Utilization** | 60% of features used | 75% of features used | 85% of features used |
| **Customer Satisfaction** | > 4.0/5.0 | > 4.2/5.0 | > 4.5/5.0 |

## 🔄 Integration Opportunities (Future)

- **Payment Processors** — Integration with Square, Clover for in-shop payments
- **Parts Suppliers** — API connections to parts distributors for inventory/ordering
- **Accounting Software** — QuickBooks/Xero integration for financial tracking
- **Marketing Tools** — Campaign management for customer retention
- **Digital Vehicle Inspection** — Photo/video documentation of vehicle issues

## 🛣️ Roadmap (Post-MVP)

| Quarter | Key Initiatives |
|---------|----------------|
| **Q2 2025** | • Customer portal for appointment requests<br>• Enhanced reporting and analytics<br>• Digital vehicle inspections with photo uploads |
| **Q3 2025** | • Parts inventory management<br>• QuickBooks/Xero integration<br>• Technician mobile companion app |
| **Q4 2025** | • Online booking widget for shop websites<br>• Multi-location support<br>• Advanced customer loyalty program |
| **Q1 2026** | • Business intelligence dashboard<br>• White-label mobile app option<br>• Integration marketplace |

## ✅ Why This Product Works

- **Solves a real pain point:** Replaces disorganized, manual processes with digital efficiency
- **Targeted niche market:** Focused on underserved small repair shops, not competing with enterprise solutions
- **Low barrier to adoption:** Simple onboarding with minimal training required
- **Clear value proposition:** Measurable improvements in customer retention and operational efficiency
- **Sustainable business model:** Predictable subscription revenue with controlled operating costs
- **Solo-founder friendly:** Intentionally designed for low maintenance and support burden
- **Scalable architecture:** Can grow from 50 to 5,000 customers without major infrastructure changes
- **Differentiated offering:** Simpler and more affordable than bloated competitors

## 🛠️ Technical Considerations

- **Data Security:** HTTPS-only with encryption at rest for all customer data
- **Backups:** Daily automated backups with 30-day retention
- **Performance:** < 1 second page load times for critical workflows
- **Availability:** 99.9% uptime target with monitoring and alerting
- **Compliance:** GDPR-ready data handling even for US customers
- **API Design:** RESTful API with future webhook capabilities
- **Disaster Recovery:** Cross-region backup strategy with documented recovery procedures