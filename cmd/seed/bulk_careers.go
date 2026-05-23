package main

import "fmt"

func bulkCareers() []careerSeed {
	// Helper: generate a career quickly with template description/roadmap
	type catDef struct {
		name string
		slug string
		desc string
		icon string
	}
	type careerDef struct {
		cat            catDef
		title, summary string
		skills         []string
		salMin, salMax int
		diff, future   int
		tags           []string
		studyDur       string
		degree         string
		workLife       int
		creative       int
		technical      int
		freelance      int
		remote         bool
		govt           bool
	}

	// Template descriptions by category style
	templates := map[string]struct {
		tasks   []string
		edu     string
		outlook string
	}{
		"technology-it": {
			tasks:   []string{"Write and test code for projects", "Collaborate with team members", "Debug and fix technical issues", "Learn new tools and frameworks", "Review code and document work", "Attend standup meetings", "Deploy and monitor applications"},
			edu:     "Bachelor's degree in Computer Science or related field, or equivalent experience. Many professionals are self-taught.",
			outlook: "Technology sector is growing rapidly in Nepal. Demand for digital skills continues to rise with remote work opportunities expanding.",
		},
		"healthcare": {
			tasks:   []string{"Examine and assess patients", "Administer treatments and medications", "Maintain patient records", "Coordinate with care team", "Educate patients and families", "Monitor vital signs", "Follow safety protocols"},
			edu:     "Bachelor's degree in relevant field. Clinical training and licensing required.",
			outlook: "Healthcare demand in Nepal is growing with population needs. Government investing in health infrastructure.",
		},
		"arts-design": {
			tasks:   []string{"Create visual content and designs", "Work with clients on creative briefs", "Iterate based on feedback", "Stay current with design trends", "Organize and manage creative assets", "Present concepts to stakeholders", "Meet project deadlines"},
			edu:     "Portfolio and practical skills matter more than formal education. Relevant training or degree helpful.",
			outlook: "Creative industries growing with digital media. Freelance opportunities expanding through global platforms.",
		},
		"media-communication": {
			tasks:   []string{"Research and develop content", "Write and edit materials", "Manage social media platforms", "Engage with audience", "Track and report metrics", "Collaborate with creative team", "Stay updated on trends"},
			edu:     "Bachelor's degree in Mass Communication, Journalism, or related field.",
			outlook: "Media landscape in Nepal is expanding with digital platforms creating new opportunities.",
		},
		"business-finance": {
			tasks:   []string{"Analyze financial data and reports", "Prepare documents and presentations", "Meet with clients or stakeholders", "Manage budgets and forecasts", "Ensure compliance with regulations", "Research market trends", "Process transactions and records"},
			edu:     "Bachelor's degree in Business, Finance, Economics, or related field.",
			outlook: "Business sector in Nepal continues to grow. Financial literacy and business skills remain in high demand.",
		},
		"skilled-trades": {
			tasks:   []string{"Read blueprints and technical diagrams", "Use hand and power tools", "Install, maintain, or repair systems", "Follow safety regulations", "Inspect work for quality", "Order materials and supplies", "Communicate with clients"},
			edu:     "Vocational training or apprenticeship. On-the-job experience valued.",
			outlook: "Skilled trades are always in demand. Cannot be outsourced. Growing with construction and infrastructure.",
		},
		"education": {
			tasks:   []string{"Plan and deliver lessons", "Assess student progress", "Prepare learning materials", "Communicate with parents/guardians", "Attend staff meetings", "Maintain classroom environment", "Participate in professional development"},
			edu:     "Bachelor's degree in Education or relevant subject. Teaching certification required.",
			outlook: "Education sector in Nepal needs qualified teachers. Government hiring and private schools expanding.",
		},
		"agriculture": {
			tasks:   []string{"Plan and manage cultivation", "Monitor crop or livestock health", "Manage resources and supplies", "Keep records and track production", "Implement sustainable practices", "Coordinate with buyers or markets", "Maintain equipment"},
			edu:     "Degree in Agriculture or related field. Practical experience highly valued.",
			outlook: "Agriculture remains Nepal's largest sector. Modern techniques and agribusiness creating new opportunities.",
		},
		"tourism-hospitality": {
			tasks:   []string{"Greet and assist customers or guests", "Manage bookings and reservations", "Coordinate services and activities", "Handle inquiries and complaints", "Maintain cleanliness and safety", "Process payments", "Promote services"},
			edu:     "Hospitality Management degree or vocational training. Language skills important.",
			outlook: "Tourism is a major industry in Nepal. Recovery and growth expected as global travel increases.",
		},
		"government-social": {
			tasks:   []string{"Process applications and documents", "Assist citizens with services", "Attend meetings and briefings", "Prepare reports and proposals", "Coordinate with agencies", "Implement programs", "Maintain records"},
			edu:     "Bachelor's degree in relevant field. Civil service exam required for government positions.",
			outlook: "Government sector in Nepal provides stable employment with benefits. Competitive entry through exams.",
		},
		"legal": {
			tasks:   []string{"Research laws and precedents", "Prepare legal documents", "Advise clients on legal matters", "Represent clients in proceedings", "Negotiate settlements", "Stay current with legal changes", "Manage case files"},
			edu:     "Bachelor of Laws (LLB) and bar exam certification required.",
			outlook: "Legal profession in Nepal growing with economic development. Specialized areas like corporate law expanding.",
		},
	}

	cat := func(name, slug, desc, icon string) catDef {
		return catDef{name, slug, desc, icon}
	}

	cd := func(c catDef, title, summary string, skills []string, salMin, salMax, diff, future int, tags []string) careerDef {
		// Scale salaries to Nepal reality: divide by ~2.5 to convert global estimates to NPR
		salMin = salMin * 2 / 5
		salMax = salMax * 2 / 5
		if salMin < 5 {
			salMin = 5
		}
		if salMax < salMin {
			salMax = salMin + 10000
		}
		return careerDef{cat: c, title: title, summary: summary, skills: skills, salMin: salMin, salMax: salMax, diff: diff, future: future, tags: tags, studyDur: "", degree: "", workLife: 3, creative: 3, technical: 3, freelance: 1}
	}

	// Categories
	health := cat("Healthcare", "healthcare", "Careers focused on helping people stay healthy, treating illness, and caring for patients.", "🏥")
	creative := cat("Arts & Design", "arts-design", "Creative careers using art, color, and design to communicate ideas visually.", "🎨")
	media := cat("Media & Communication", "media-communication", "Careers in journalism, broadcasting, content creation, and digital media.", "📡")
	biz := cat("Business & Finance", "business-finance", "Careers managing money, tracking finances, and helping businesses succeed.", "💰")
	trades := cat("Skilled Trades", "skilled-trades", "Hands-on jobs requiring technical skills, learned through apprenticeship.", "🔧")
	edu := cat("Education", "education", "Careers dedicated to teaching and helping students learn and grow.", "📚")
	agri := cat("Agriculture & Environment", "agriculture-environment", "Careers in farming, environmental conservation, and natural resource management.", "🌱")
	tourism := cat("Tourism & Hospitality", "tourism-hospitality", "Careers serving travelers, tourists, and guests in Nepal's hospitality industry.", "🏔️")
	govt := cat("Government & Social Services", "government-social", "Careers in public service, civil service, and social work.", "🏛️")
	legalCat := cat("Legal", "legal", "Careers in law, justice, and legal services.", "⚖️")
	digital := cat("Digital & Modern Careers", "digital-modern", "Modern digital careers including content creation, freelancing, and gig economy.", "📱")
	fitness := cat("Fitness & Sports", "fitness-sports", "Careers in physical fitness, sports coaching, and athletic training.", "💪")
	music := cat("Music & Entertainment", "music-entertainment", "Careers in music production, performance, and entertainment.", "🎵")
	transport := cat("Transportation & Logistics", "transportation-logistics", "Careers moving goods and people efficiently.", "🚚")
	manufacturing := cat("Manufacturing & Industry", "manufacturing-industry", "Careers in production, manufacturing, and industrial operations.", "🏭")
	science := cat("Science & Research", "science-research", "Careers in scientific research, laboratory work, and discovery.", "🔬")
	realestate := cat("Real Estate & Construction", "real-estate-construction", "Careers in property, construction, and infrastructure development.", "🏗️")
	beauty := cat("Beauty & Wellness", "beauty-wellness", "Careers in personal care, beauty services, and wellness.", "💆")
	emerging := cat("Emerging & Future Careers", "emerging-future", "New and emerging career paths in cutting-edge fields.", "🚀")
	vocation := cat("Vocational & Technical", "vocational-technical", "Practical career paths requiring specific technical training.", "🛠️")
	ecommerce := cat("E-Commerce & Retail", "ecommerce-retail", "Careers in online and offline retail, sales, and commerce.", "🛒")

	defs := []careerDef{
		// === DIGITAL & MODERN CAREERS (40+) ===
		cd(digital, "YouTuber", "Create video content for YouTube, build an audience, and earn through ads, sponsorships, and merchandise.", []string{"Video production and editing", "Content planning and scripting", "Audience engagement", "SEO for YouTube", "Camera presence", "Thumbnail design", "Analytics tracking"}, 30000, 200000, 2, 55, []string{"content-creation", "video", "social-media", "creative", "freelance"}),
		cd(digital, "Thumbnail Designer", "Design eye-catching YouTube thumbnails that drive clicks and views for content creators.", []string{"Photoshop/GIMP skills", "Visual composition", "Typography", "Color theory", "Understanding of click psychology", "Quick turnaround", "Brand consistency"}, 25000, 80000, 2, 60, []string{"design", "content-creation", "freelance", "remote-work", "creative"}),
		cd(digital, "Content Creator", "Create engaging content across social media platforms including Instagram, TikTok, and Facebook.", []string{"Content strategy", "Video shooting and editing", "Social media trends", "Audience growth", "Brand collaboration", "Analytics", "Storytelling"}, 20000, 120000, 2, 60, []string{"content-creation", "social-media", "creative", "freelance", "digital"}),
		cd(digital, "Motion Graphics Artist", "Create animated graphics, visual effects, and motion design for videos, ads, and presentations.", []string{"After Effects", "Animation principles", "3D basics", "Typography animation", "Visual storytelling", "Client communication", "Project management"}, 35000, 150000, 3, 65, []string{"animation", "design", "creative", "freelance", "remote-work"}),
		cd(digital, "Streamer", "Live stream gaming, creative content, or talk shows on platforms like Twitch and YouTube.", []string{"Live streaming setup", "Audience engagement", "Consistent scheduling", "Basic video/audio tech", "Community management", "Content planning", "Sponsorship outreach"}, 15000, 200000, 2, 45, []string{"entertainment", "gaming", "freelance", "social-media", "creative"}),
		cd(digital, "Social Media Manager", "Manage social media accounts for brands, create content calendars, and grow online presence.", []string{"Platform expertise (Insta, TikTok, FB, LinkedIn)", "Content creation", "Analytics and reporting", "Community management", "Copywriting", "Paid social ads", "Trend awareness"}, 30000, 100000, 3, 70, []string{"social-media", "marketing", "digital", "remote-work", "creative"}),
		cd(digital, "UI/UX Designer", "Design user interfaces and experiences for websites, apps, and digital products.", []string{"Figma/Sketch/Adobe XD", "User research", "Wireframing and prototyping", "Visual design", "Usability testing", "Interaction design", "Design systems"}, 40000, 180000, 3, 80, []string{"design", "tech", "creative", "remote-work", "high-growth"}),
		cd(digital, "Ethical Hacker", "Find security vulnerabilities in systems and networks to help organizations improve their security.", []string{"Network security", "Penetration testing", "Python programming", "Security tools (Metasploit, Wireshark)", "Linux expertise", "Social engineering awareness", "Report writing"}, 50000, 250000, 5, 90, []string{"tech", "security", "high-growth", "remote-work", "high-salary"}),
		cd(digital, "Cloud Engineer", "Design and manage cloud infrastructure on AWS, Google Cloud, or Azure.", []string{"Cloud platforms (AWS/GCP/Azure)", "Infrastructure as Code", "Containerization (Docker/K8s)", "Networking", "Security best practices", "Automation", "Monitoring and logging"}, 60000, 250000, 4, 92, []string{"tech", "cloud", "high-growth", "remote-work", "high-salary"}),
		cd(digital, "Game Developer", "Design and build video games for PC, console, or mobile platforms.", []string{"Game engine (Unity/Unreal)", "Programming (C#/C++)", "Game design principles", "3D modeling basics", "Problem-solving", "Team collaboration", "Version control"}, 30000, 150000, 4, 70, []string{"tech", "gaming", "creative", "remote-work", "engineering"}),
		cd(digital, "AI Engineer", "Build and deploy artificial intelligence and machine learning models for real-world applications.", []string{"Python", "Machine learning frameworks", "Deep learning", "Data processing", "Model deployment", "Statistics", "Problem-solving"}, 70000, 300000, 5, 95, []string{"tech", "ai", "high-growth", "high-salary", "future-proof"}),
		cd(digital, "Prompt Engineer", "Craft and optimize prompts for AI language models to get the best outputs.", []string{"Understanding of LLMs", "Creative writing", "Logical reasoning", "Experimentation mindset", "Domain knowledge", "Pattern recognition", "Technical writing"}, 40000, 180000, 3, 85, []string{"tech", "ai", "remote-work", "emerging", "creative"}),
		cd(digital, "Robotics Technician", "Install, maintain, and repair robotic systems in manufacturing and industrial settings.", []string{"Mechanical aptitude", "Electronics basics", "Programming (PLC/Robotics)", "Troubleshooting", "Blueprint reading", "Safety protocols", "Precision measurement"}, 35000, 120000, 4, 85, []string{"tech", "robotics", "hands-on", "high-growth", "engineering"}),
		cd(digital, "Cybersecurity Analyst", "Protect organizations from cyber threats by monitoring, detecting, and responding to security incidents.", []string{"Security monitoring", "Threat detection", "Incident response", "Risk assessment", "Security tools (SIEM, IDS/IPS)", "Network security", "Compliance knowledge"}, 45000, 200000, 4, 92, []string{"tech", "security", "high-growth", "remote-work", "high-salary"}),
		cd(digital, "SEO Specialist", "Optimize websites to rank higher in search engine results and drive organic traffic.", []string{"Keyword research", "On-page optimization", "Technical SEO", "Link building", "Analytics (Google Analytics)", "Content strategy", "SEO tools (Ahrefs, SEMrush)"}, 25000, 90000, 3, 68, []string{"digital-marketing", "tech", "remote-work", "high-growth", "freelance"}),
		cd(digital, "TikTok Creator", "Create short-form video content for TikTok, build followers, and monetize through brand deals and creator fund.", []string{"Video editing", "Trend awareness", "Storytelling", "Camera presence", "Analytics", "Community engagement", "Content planning"}, 15000, 150000, 2, 50, []string{"social-media", "content-creation", "creative", "freelance", "digital"}),
		cd(digital, "Voice-Over Artist", "Provide voice recordings for commercials, animations, audiobooks, and video content.", []string{"Voice control and modulation", "Script interpretation", "Recording setup", "Audio editing basics", "Character voices", "Accent and dialect work", "Client communication"}, 20000, 80000, 2, 60, []string{"creative", "media", "freelance", "remote-work", "entertainment"}),
		cd(digital, "Freelancer", "Work independently offering skills and services to clients on platforms like Upwork, Fiverr, and Freelancer.", []string{"Self-discipline", "Time management", "Client communication", "Specific skill (writing, design, code, etc.)", "Proposal writing", "Financial management", "Personal branding"}, 15000, 150000, 2, 80, []string{"freelance", "remote-work", "digital", "flexible", "entrepreneur"}),
		cd(digital, "Copywriter", "Write persuasive marketing copy for websites, emails, ads, and social media.", []string{"Persuasive writing", "SEO copywriting", "Research skills", "Understanding of marketing", "Adaptability in tone", "Editing and proofreading", "Client management"}, 25000, 90000, 2, 65, []string{"writing", "marketing", "creative", "freelance", "remote-work"}),
		cd(digital, "Blogger", "Write and publish blog content on topics of interest, monetizing through ads, affiliates, and sponsored content.", []string{"Writing and editing", "SEO", "Niche expertise", "Social media promotion", "Basic web design", "Email marketing", "Audience building"}, 15000, 70000, 2, 55, []string{"writing", "content-creation", "freelance", "remote-work", "creative"}),
		cd(digital, "Digital Marketing Manager", "Plan and execute digital marketing strategies across channels to drive business growth.", []string{"Strategy development", "Social media marketing", "Email marketing", "PPC advertising", "Analytics and reporting", "Content marketing", "Team leadership"}, 40000, 150000, 3, 75, []string{"marketing", "digital", "high-growth", "business", "creative"}),
		cd(digital, "E-commerce Manager", "Manage online stores, optimize product listings, and drive sales through digital channels.", []string{"Platform expertise (Shopify/WooCommerce)", "Product listing optimization", "Digital advertising", "Inventory management", "Customer service", "Analytics", "Conversion optimization"}, 35000, 120000, 3, 72, []string{"ecommerce", "business", "digital", "high-growth", "marketing"}),

		// === CREATIVE ARTS & DESIGN (30+) ===
		cd(creative, "Animator", "Create animated content for films, games, advertisements, and digital media.", []string{"Animation software (Blender, Maya, After Effects)", "Timing and motion", "Storyboarding", "Character design", "Visual storytelling", "Attention to detail", "Creativity"}, 30000, 120000, 3, 60, []string{"animation", "creative", "design", "freelance", "remote-work"}),
		cd(creative, "3D Artist", "Create three-dimensional models, textures, and environments for games, films, and VR.", []string{"3D modeling (Blender, Maya, 3ds Max)", "Texturing and shading", "Lighting and rendering", "Sculpting basics", "UV mapping", "PBR workflows", "Portfolio management"}, 35000, 130000, 4, 65, []string{"3d", "creative", "design", "freelance", "tech"}),
		cd(creative, "Photographer", "Capture professional photos for events, portraits, products, and commercial use.", []string{"Camera operation", "Lighting techniques", "Photo editing (Lightroom/Photoshop)", "Composition", "Client management", "Business skills", "Equipment knowledge"}, 20000, 100000, 2, 50, []string{"photography", "creative", "freelance", "art", "events"}),
		cd(creative, "Interior Designer", "Design functional and aesthetic interior spaces for homes, offices, and commercial buildings.", []string{"Design software (AutoCAD, SketchUp)", "Space planning", "Color theory", "Material knowledge", "Client consultation", "Project management", "Trend awareness"}, 30000, 120000, 3, 60, []string{"design", "creative", "real-estate", "freelance", "art"}),
		cd(creative, "Fashion Designer", "Design clothing and accessories, create patterns, and oversee production.", []string{"Sewing and pattern-making", "Fabric knowledge", "Fashion illustration", "Trend forecasting", "Product development", "Brand building", "Business acumen"}, 25000, 100000, 4, 50, []string{"fashion", "creative", "design", "art", "entrepreneur"}),
		cd(creative, "Tattoo Artist", "Create permanent body art designs and apply tattoos professionally.", []string{"Drawing and design", "Sterilization and safety", "Tattoo machine operation", "Color theory", "Client consultation", "Artistic skill", "Hygiene protocols"}, 30000, 150000, 3, 55, []string{"art", "creative", "freelance", "hands-on", "beauty"}),
		cd(creative, "Calligrapher", "Create decorative handwriting and lettering for invitations, logos, and art pieces.", []string{"Hand lettering", "Pen and brush control", "Composition", "Ink and paper knowledge", "Digital calligraphy", "Client communication", "Business management"}, 15000, 60000, 2, 40, []string{"art", "creative", "freelance", "design", "handmade"}),
		cd(creative, "Illustrator", "Create original illustrations for books, magazines, advertisements, and digital media.", []string{"Drawing and painting", "Digital illustration tools", "Style development", "Visual storytelling", "Color theory", "Client collaboration", "Portfolio management"}, 20000, 90000, 2, 55, []string{"art", "illustration", "creative", "freelance", "remote-work"}),
		cd(creative, "Jewelry Designer", "Design and create jewelry pieces using various materials and techniques.", []string{"Design sketching", "Metalworking", "Gemstone knowledge", "CAD for jewelry", "Attention to detail", "Creativity", "Business skills"}, 20000, 80000, 3, 45, []string{"design", "creative", "handmade", "art", "business"}),

		// === MEDIA & COMMUNICATION (25+) ===
		cd(media, "Podcaster", "Create and host audio content on topics of expertise, building audience and monetizing.", []string{"Audio recording and editing", "Interview skills", "Content planning", "Storytelling", "Audience building", "Sponsorship outreach", "Consistent publishing"}, 10000, 100000, 2, 60, []string{"media", "content-creation", "freelance", "creative", "digital"}),
		cd(media, "Video Editor", "Edit raw footage into polished video content for YouTube, films, commercials, and social media.", []string{"Editing software (Premiere Pro, DaVinci Resolve)", "Color grading", "Sound design", "Storytelling through editing", "Timing and pacing", "Motion graphics basics", "Client management"}, 25000, 100000, 3, 65, []string{"video", "creative", "freelance", "remote-work", "media"}),
		cd(media, "Radio Jockey", "Host radio shows, play music, conduct interviews, and engage with listeners.", []string{"Voice modulation", "Script reading", "Music knowledge", "Live interview skills", "Audience engagement", "Technical audio skills", "Time management"}, 20000, 70000, 2, 45, []string{"media", "entertainment", "communication", "creative", "broadcasting"}),
		cd(media, "News Anchor", "Present news on television or digital platforms, delivering information clearly and professionally.", []string{"On-camera presence", "Script delivery", "News judgment", "Research skills", "Interview techniques", "Ad-lib ability", "Professional appearance"}, 35000, 150000, 3, 55, []string{"media", "journalism", "communication", "broadcasting", "prestigious"}),
		cd(media, "Journalist", "Research, investigate, and report news stories for print, broadcast, or digital media.", []string{"Writing and reporting", "Research and fact-checking", "Interview skills", "Digital media tools", "Ethics and objectivity", "Deadline management", "Source development"}, 20000, 80000, 3, 50, []string{"media", "journalism", "writing", "communication", "research"}),
		cd(media, "Content Writer", "Write articles, blog posts, website content, and marketing materials for various clients.", []string{"Excellent writing skills", "SEO knowledge", "Research ability", "Adaptability in tone", "Editing and proofreading", "Time management", "Topic expertise"}, 20000, 70000, 2, 60, []string{"writing", "content", "freelance", "remote-work", "digital"}),
		cd(media, "Editor", "Review and refine written content for clarity, accuracy, style, and grammar.", []string{"Grammar and style mastery", "Attention to detail", "Fact-checking", "Structural editing", "Collaboration with writers", "Subject matter expertise", "Deadline management"}, 25000, 80000, 3, 55, []string{"writing", "media", "editing", "communication", "publishing"}),
		cd(media, "Public Relations Officer", "Manage public image and communication for organizations, handling media relations and crisis communication.", []string{"Media relations", "Press release writing", "Crisis management", "Event planning", "Social media management", "Communication strategy", "Stakeholder management"}, 30000, 100000, 3, 60, []string{"media", "communication", "business", "corporate", "marketing"}),

		// === BUSINESS & FINANCE (30+) ===
		cd(biz, "Financial Analyst", "Analyze financial data to help businesses and individuals make investment decisions.", []string{"Financial modeling", "Data analysis", "Excel expertise", "Market research", "Report writing", "Valuation techniques", "Attention to detail"}, 40000, 150000, 4, 72, []string{"finance", "business", "analytics", "high-salary", "corporate"}),
		cd(biz, "Business Consultant", "Advise organizations on strategy, operations, and management to improve performance.", []string{"Problem-solving", "Strategic thinking", "Data analysis", "Client management", "Presentation skills", "Industry expertise", "Project management"}, 50000, 200000, 4, 70, []string{"business", "consulting", "high-salary", "corporate", "strategy"}),
		cd(biz, "Entrepreneur", "Start and grow your own business, creating value and jobs in the community.", []string{"Business planning", "Leadership", "Financial management", "Marketing", "Networking", "Problem-solving", "Resilience"}, 0, 500000, 5, 60, []string{"business", "entrepreneur", "leadership", "creative", "high-risk"}),
		cd(biz, "Human Resources Manager", "Manage employee relations, recruitment, training, and organizational culture.", []string{"Recruitment and hiring", "Employee relations", "Labor law knowledge", "Training and development", "Communication", "Conflict resolution", "Organizational skills"}, 35000, 120000, 3, 65, []string{"business", "hr", "corporate", "stable", "people"}),
		cd(biz, "Project Manager", "Plan, execute, and close projects, coordinating teams and resources to meet goals.", []string{"Project planning", "Team leadership", "Risk management", "Stakeholder communication", "Budget management", "Agile/Scrum knowledge", "Problem-solving"}, 40000, 150000, 3, 70, []string{"business", "management", "corporate", "high-growth", "leadership"}),
		cd(biz, "Supply Chain Manager", "Oversee product flow from raw materials to delivery, optimizing efficiency and cost.", []string{"Logistics management", "Inventory control", "Supplier relationship", "Data analysis", "Process optimization", "Negotiation", "ERP systems knowledge"}, 35000, 130000, 3, 72, []string{"business", "logistics", "management", "stable", "operations"}),
		cd(biz, "Market Research Analyst", "Study market conditions to identify potential sales opportunities and consumer preferences.", []string{"Research methodology", "Data analysis", "Statistical software", "Report writing", "Survey design", "Consumer behavior understanding", "Presentation skills"}, 30000, 100000, 3, 68, []string{"business", "analytics", "marketing", "research", "corporate"}),
		cd(biz, "Investment Banker", "Help organizations raise capital and advise on mergers, acquisitions, and financial strategy.", []string{"Financial modeling", "Valuation", "Negotiation", "Client relationship management", "Deal execution", "Market analysis", "Presentation skills"}, 80000, 350000, 5, 68, []string{"finance", "high-salary", "business", "prestigious", "corporate"}),
		cd(biz, "Sales Manager", "Lead sales teams, set targets, develop strategies, and drive revenue growth.", []string{"Sales strategy", "Team leadership", "Client relationship", "Forecasting", "Negotiation", "CRM tools", "Communication"}, 35000, 150000, 3, 60, []string{"business", "sales", "management", "high-growth", "corporate"}),

		// === HEALTHCARE (30+) ===
		cd(health, "Medical Lab Technician", "Perform laboratory tests to help diagnose and treat diseases.", []string{"Lab equipment operation", "Sample collection", "Test analysis", "Quality control", "Attention to detail", "Safety protocols", "Record keeping"}, 25000, 70000, 3, 75, []string{"healthcare", "lab", "stable", "hands-on", "technical"}),
		cd(health, "Radiologist", "Interpret medical images like X-rays, CT scans, and MRIs to diagnose conditions.", []string{"Image interpretation", "Radiology equipment knowledge", "Anatomy expertise", "Diagnostic skills", "Patient communication", "Report writing", "Attention to detail"}, 80000, 300000, 5, 82, []string{"healthcare", "medical", "high-salary", "specialist", "prestigious"}),
		cd(health, "Nutritionist", "Advise clients on diet, nutrition, and healthy eating habits for better health.", []string{"Nutrition science", "Client assessment", "Meal planning", "Health coaching", "Communication", "Research knowledge", "Behavioral change"}, 20000, 60000, 2, 65, []string{"healthcare", "wellness", "consulting", "freelance", "helping-people"}),
		cd(health, "Optometrist", "Examine eyes, prescribe glasses and contact lenses, and detect eye diseases.", []string{"Eye examination", "Refraction", "Contact lens fitting", "Eye disease detection", "Patient education", "Equipment operation", "Record keeping"}, 40000, 120000, 4, 72, []string{"healthcare", "medical", "stable", "clinical", "helping-people"}),
		cd(health, "Veterinarian", "Diagnose and treat illnesses in animals, from pets to livestock.", []string{"Animal diagnosis", "Surgical skills", "Vaccination", "Client communication", "Medical record keeping", "Compassion for animals", "Practice management"}, 30000, 100000, 4, 75, []string{"healthcare", "animals", "medical", "stable", "helping-people"}),
		cd(health, "Dental Hygienist", "Clean teeth, examine patients for oral diseases, and provide preventive dental care.", []string{"Teeth cleaning", "X-ray operation", "Patient education", "Oral health assessment", "Instrument sterilization", "Record keeping", "Communication"}, 25000, 70000, 3, 78, []string{"healthcare", "dental", "stable", "clinical", "helping-people"}),
		cd(health, "Paramedic", "Respond to medical emergencies, provide pre-hospital care, and transport patients.", []string{"Emergency assessment", "Life support", "Wound management", "Patient transport", "Quick decision-making", "Stress management", "Teamwork"}, 20000, 60000, 4, 80, []string{"healthcare", "emergency", "hands-on", "helping-people", "high-stress"}),

		// === FITNESS & SPORTS (15+) ===
		cd(fitness, "Fitness Influencer", "Build a following around fitness content, sharing workouts, nutrition tips, and lifestyle advice.", []string{"Social media content", "Fitness knowledge", "Video production", "Audience engagement", "Personal branding", "Sponsorship management", "Community building"}, 15000, 150000, 2, 55, []string{"fitness", "social-media", "influencer", "creative", "freelance"}),
		cd(fitness, "Yoga Teacher", "Teach yoga classes, guide meditation, and help students improve flexibility and mindfulness.", []string{"Yoga poses and sequences", "Breathing techniques", "Meditation guidance", "Class management", "Anatomy knowledge", "Communication", "Empathy"}, 15000, 60000, 2, 70, []string{"fitness", "wellness", "teaching", "freelance", "helping-people"}),
		cd(fitness, "Personal Trainer", "Design and deliver personalized fitness programs to help clients achieve health goals.", []string{"Exercise programming", "Nutrition basics", "Client motivation", "Anatomy knowledge", "Safety and form", "Business skills", "Communication"}, 20000, 80000, 2, 65, []string{"fitness", "health", "freelance", "helping-people", "hands-on"}),
		cd(fitness, "Sports Coach", "Train athletes and teams in specific sports, developing skills and strategies.", []string{"Sport-specific expertise", "Training program design", "Motivation techniques", "Game strategy", "Player development", "Communication", "Leadership"}, 20000, 80000, 3, 60, []string{"sports", "coaching", "teaching", "leadership", "hands-on"}),
		cd(fitness, "Physical Therapist", "Help patients recover from injuries and improve mobility through targeted exercises.", []string{"Therapy techniques", "Patient assessment", "Exercise prescription", "Manual therapy", "Pain management", "Progress tracking", "Patient education"}, 30000, 100000, 4, 78, []string{"healthcare", "fitness", "therapy", "helping-people", "clinical"}),
		cd(fitness, "Sports Nutritionist", "Advise athletes on diet and nutrition to optimize performance and recovery.", []string{"Sports nutrition science", "Meal planning", "Supplement knowledge", "Hydration strategies", "Body composition analysis", "Education and coaching", "Research"}, 25000, 80000, 3, 70, []string{"fitness", "nutrition", "health", "sports", "consulting"}),

		// === MUSIC & ENTERTAINMENT (15+) ===
		cd(music, "Music Producer", "Produce and record music tracks, working with artists on sound design and mixing.", []string{"DAW expertise (Ableton, FL Studio)", "Sound design", "Mixing and mastering", "Music theory", "Artist collaboration", "Recording techniques", "Creativity"}, 25000, 150000, 3, 55, []string{"music", "creative", "entertainment", "freelance", "art"}),
		cd(music, "Singer", "Perform vocal music for recordings, live shows, and events.", []string{"Vocal technique", "Stage presence", "Music theory", "Performance skills", "Recording studio experience", "Repertoire building", "Networking"}, 10000, 100000, 3, 45, []string{"music", "entertainment", "creative", "performance", "freelance"}),
		cd(music, "DJ", "Mix and play recorded music for audiences at events, clubs, and festivals.", []string{"DJ equipment operation", "Music selection", "Mixing and beatmatching", "Crowd reading", "Music library management", "Technical setup", "Marketing"}, 15000, 100000, 2, 45, []string{"music", "entertainment", "events", "creative", "freelance"}),
		cd(music, "Sound Engineer", "Manage audio quality for live events, recordings, and broadcasts.", []string{"Audio equipment operation", "Mixing consoles", "Microphone techniques", "Signal processing", "Acoustics knowledge", "Problem-solving", "Teamwork"}, 20000, 80000, 3, 60, []string{"music", "tech", "events", "media", "freelance"}),
		cd(music, "Composer", "Create original music for films, games, commercials, and other media.", []string{"Music theory", "Instrument proficiency", "DAW expertise", "Orchestration", "Client collaboration", "Timing and synchronization", "Creativity"}, 20000, 120000, 4, 55, []string{"music", "creative", "media", "freelance", "art"}),
		cd(music, "Dance Instructor", "Teach dance techniques and choreography to students of all levels.", []string{"Dance expertise", "Teaching ability", "Choreography", "Patience", "Communication", "Class management", "Performance skills"}, 15000, 50000, 2, 50, []string{"dance", "fitness", "teaching", "creative", "freelance"}),

		// === EMERGING & FUTURE CAREERS (30+) ===
		cd(emerging, "Data Analyst", "Analyze data to help organizations make better decisions through insights and reporting.", []string{"SQL", "Excel/Google Sheets", "Data visualization", "Statistics basics", "Python or R basics", "Critical thinking", "Communication"}, 30000, 120000, 3, 82, []string{"tech", "analytics", "high-growth", "remote-work", "data"}),
		cd(emerging, "Machine Learning Engineer", "Design and implement machine learning systems that learn from data.", []string{"Python", "ML frameworks (TensorFlow, PyTorch)", "Statistics", "Data engineering", "Model deployment", "Software engineering", "Research skills"}, 60000, 250000, 5, 95, []string{"tech", "ai", "high-salary", "high-growth", "future-proof"}),
		cd(emerging, "Blockchain Developer", "Build decentralized applications and smart contracts on blockchain platforms.", []string{"Solidity", "Web3.js", "Smart contract security", "Distributed systems", "JavaScript/Python", "Cryptography basics", "DeFi knowledge"}, 50000, 250000, 5, 70, []string{"tech", "blockchain", "high-salary", "remote-work", "emerging"}),
		cd(emerging, "IoT Engineer", "Design and implement Internet of Things systems connecting devices and sensors.", []string{"Embedded systems", "Sensor integration", "Network protocols", "Cloud platforms", "Data processing", "Hardware knowledge", "Security"}, 40000, 180000, 4, 85, []string{"tech", "iot", "engineering", "high-growth", "emerging"}),
		cd(emerging, "AR/VR Developer", "Create augmented and virtual reality experiences for gaming, training, and education.", []string{"Unity/Unreal Engine", "3D modeling", "C#/C++", "Spatial computing", "UI/UX for XR", "Performance optimization", "Creativity"}, 40000, 180000, 4, 78, []string{"tech", "ar-vr", "creative", "emerging", "high-growth"}),
		cd(emerging, "Quantum Computing Engineer", "Research and develop quantum computing algorithms and hardware.", []string{"Quantum mechanics", "Linear algebra", "Programming (Qiskit, Cirq)", "Algorithm design", "Physics background", "Mathematics", "Research methodology"}, 80000, 350000, 5, 95, []string{"tech", "quantum", "research", "high-salary", "future-proof"}),
		cd(emerging, "Bioinformatician", "Analyze biological data using computational tools and techniques.", []string{"Bioinformatics tools", "Programming (Python/R)", "Statistics", "Genomics knowledge", "Data analysis", "Biology background", "Research methods"}, 40000, 150000, 4, 85, []string{"science", "bio", "tech", "research", "high-growth"}),
		cd(emerging, "Sustainability Consultant", "Advise organizations on environmental sustainability and green practices.", []string{"Environmental science", "Sustainability frameworks", "Energy efficiency", "Waste management", "Client consulting", "Report writing", "Regulation knowledge"}, 30000, 120000, 3, 80, []string{"environment", "consulting", "sustainability", "green", "high-growth"}),
		cd(emerging, "Drone Pilot", "Operate drones for aerial photography, surveying, agriculture, and delivery services.", []string{"Drone operation", "Flight regulations", "Aerial photography", "Maintenance", "Mission planning", "Safety protocols", "Basic photography"}, 25000, 90000, 3, 65, []string{"tech", "aviation", "outdoor", "hands-on", "emerging"}),
		cd(emerging, "EV Mechanic", "Repair and maintain electric vehicles, batteries, and charging systems.", []string{"EV systems knowledge", "High-voltage safety", "Diagnostic tools", "Battery maintenance", "Electrical engineering basics", "Customer service", "Problem-solving"}, 30000, 100000, 3, 88, []string{"automotive", "tech", "green", "hands-on", "high-growth"}),
		cd(emerging, "Solar Technician", "Install and maintain solar panel systems for homes and businesses.", []string{"Solar panel installation", "Electrical wiring", "Roof work", "System maintenance", "Safety protocols", "Customer communication", "Basic electrical knowledge"}, 25000, 80000, 3, 85, []string{"green", "energy", "tech", "hands-on", "high-growth"}),
		cd(emerging, "Wind Turbine Technician", "Install, maintain, and repair wind turbines at wind energy farms.", []string{"Mechanical skills", "Electrical knowledge", "Climbing and safety", "Turbine diagnostics", "Hydraulic systems", "Troubleshooting", "Physical fitness"}, 30000, 100000, 4, 82, []string{"green", "energy", "tech", "hands-on", "high-growth"}),

		// === SKILLED TRADES (30+) ===
		cd(trades, "HVAC Technician", "Install and repair heating, ventilation, and air conditioning systems.", []string{"HVAC system knowledge", "Refrigeration", "Electrical troubleshooting", "Pipe fitting", "Customer service", "Safety compliance", "Diagnostic skills"}, 25000, 80000, 3, 80, []string{"trades", "hands-on", "tech", "stable", "no-college"}),
		cd(trades, "Carpenter", "Build and install wooden structures, furniture, and frameworks for buildings.", []string{"Woodworking", "Blueprint reading", "Power tool operation", "Measuring and marking", "Joinery techniques", "Physical stamina", "Attention to detail"}, 20000, 70000, 3, 65, []string{"trades", "hands-on", "construction", "no-college", "stable"}),
		cd(trades, "Painter", "Apply paint and finishes to buildings, structures, and surfaces.", []string{"Painting techniques", "Surface preparation", "Color mixing", "Spray equipment", "Safety practices", "Time management", "Attention to detail"}, 15000, 50000, 2, 55, []string{"trades", "hands-on", "construction", "no-college", "stable"}),
		cd(trades, "Welder", "Join metal parts using welding techniques for construction and manufacturing.", []string{"Welding techniques", "Metal properties", "Blueprint reading", "Safety compliance", "Precision measurement", "Equipment maintenance", "Physical stamina"}, 20000, 70000, 3, 70, []string{"trades", "hands-on", "manufacturing", "no-college", "essential"}),
		cd(trades, "Auto Mechanic", "Diagnose, repair, and maintain automobiles and light trucks.", []string{"Engine diagnostics", "Brake and suspension", "Electrical systems", "Customer service", "Tool proficiency", "Problem-solving", "Physical stamina"}, 20000, 70000, 3, 65, []string{"trades", "automotive", "hands-on", "stable", "essential"}),
		cd(trades, "Locksmith", "Install, repair, and adjust locks and security systems for homes and businesses.", []string{"Lock mechanisms", "Key cutting", "Security system knowledge", "Customer service", "Precision work", "Problem-solving", "Business management"}, 20000, 60000, 2, 60, []string{"trades", "security", "hands-on", "no-college", "stable"}),
		cd(trades, "Heavy Equipment Operator", "Operate construction machinery like bulldozers, excavators, and cranes.", []string{"Equipment operation", "Safety protocols", "Basic maintenance", "Site reading", "Hand-eye coordination", "Physical stamina", "Teamwork"}, 25000, 80000, 3, 68, []string{"trades", "construction", "hands-on", "outdoor", "no-college"}),
		cd(trades, "Elevator Mechanic", "Install, maintain, and repair elevators, escalators, and moving walkways.", []string{"Mechanical systems", "Electrical troubleshooting", "Blueprint reading", "Safety compliance", "Precision work", "Customer communication", "Problem-solving"}, 30000, 100000, 4, 82, []string{"trades", "mechanical", "tech", "hands-on", "high-salary"}),
		cd(trades, "Construction Worker", "Perform general labor tasks on construction sites, supporting various trades.", []string{"Physical stamina", "Tool use", "Safety awareness", "Teamwork", "Basic construction knowledge", "Reliability", "Adaptability"}, 15000, 40000, 2, 55, []string{"trades", "construction", "hands-on", "outdoor", "no-college"}),

		// === TOURISM & HOSPITALITY (25+) ===
		cd(tourism, "Tour Guide", "Lead groups of tourists to attractions, providing information and ensuring enjoyable experiences.", []string{"Local knowledge", "Communication skills", "Language proficiency", "Group management", "Storytelling", "Customer service", "Problem-solving"}, 15000, 50000, 2, 55, []string{"tourism", "hospitality", "outdoor", "people", "flexible"}),
		cd(tourism, "Travel Agent", "Help clients plan and book travel arrangements including flights, hotels, and tours.", []string{"Booking systems", "Destination knowledge", "Customer service", "Sales skills", "Trip planning", "Problem-solving", "Technology proficiency"}, 18000, 60000, 2, 45, []string{"tourism", "travel", "office-job", "people", "sales"}),
		cd(tourism, "Restaurant Manager", "Oversee daily operations of restaurants, ensuring quality service and profitability.", []string{"Staff management", "Customer service", "Inventory control", "Financial management", "Health compliance", "Scheduling", "Problem-solving"}, 25000, 80000, 3, 55, []string{"hospitality", "management", "people", "business", "stable"}),
		cd(tourism, "Spa Manager", "Manage spa operations, staff, and services to provide excellent wellness experiences.", []string{"Spa service knowledge", "Staff management", "Customer service", "Inventory management", "Scheduling", "Wellness trends", "Business management"}, 25000, 70000, 3, 58, []string{"wellness", "hospitality", "management", "people", "beauty"}),
		cd(tourism, "Event Planner", "Plan and coordinate events like weddings, conferences, and parties.", []string{"Organization", "Vendor coordination", "Budget management", "Client communication", "Creativity", "Problem-solving", "Time management"}, 20000, 70000, 3, 58, []string{"events", "hospitality", "creative", "people", "freelance"}),
		cd(tourism, "Barista", "Prepare and serve coffee and other beverages in cafes and coffee shops.", []string{"Coffee preparation", "Customer service", "Cash handling", "Cleanliness", "Speed and efficiency", "Product knowledge", "Teamwork"}, 10000, 30000, 1, 45, []string{"hospitality", "food-service", "people", "entry-level", "flexible"}),
		cd(tourism, "Ski Instructor", "Teach skiing techniques to individuals and groups at mountain resorts.", []string{"Skiing expertise", "Teaching ability", "Safety awareness", "Physical fitness", "Patience", "Communication", "Customer service"}, 20000, 70000, 3, 45, []string{"sports", "outdoor", "teaching", "seasonal", "tourism"}),

		// === AGRICULTURE & ENVIRONMENT (20+) ===
		cd(agri, "Organic Farmer", "Grow crops using organic methods, focusing on sustainability and chemical-free production.", []string{"Farming techniques", "Soil management", "Pest control", "Crop rotation", "Marketing", "Business management", "Sustainability knowledge"}, 15000, 60000, 3, 65, []string{"agriculture", "organic", "outdoor", "sustainable", "business"}),
		cd(agri, "Agronomist", "Advise farmers on crop production, soil management, and sustainable farming practices.", []string{"Crop science", "Soil analysis", "Pest management", "Research methods", "Communication", "Problem-solving", "Technology adoption"}, 25000, 80000, 3, 72, []string{"agriculture", "science", "consulting", "outdoor", "stable"}),
		cd(agri, "Forestry Technician", "Manage forest resources, conduct surveys, and support conservation efforts.", []string{"Forest ecology", "Tree identification", "GPS and mapping", "Data collection", "Conservation practices", "Physical stamina", "Safety awareness"}, 20000, 60000, 3, 70, []string{"environment", "outdoor", "conservation", "hands-on", "stable"}),
		cd(agri, "Fisheries Manager", "Manage fish populations and aquatic resources for commercial or conservation purposes.", []string{"Aquatic biology", "Population management", "Water quality testing", "Regulation knowledge", "Record keeping", "Equipment maintenance", "Team management"}, 20000, 70000, 3, 65, []string{"environment", "agriculture", "science", "outdoor", "management"}),
		cd(agri, "Landscape Architect", "Design outdoor spaces including parks, gardens, and public areas.", []string{"Design software (AutoCAD, SketchUp)", "Plant knowledge", "Site analysis", "Environmental awareness", "Creativity", "Project management", "Client communication"}, 25000, 90000, 3, 62, []string{"design", "environment", "creative", "outdoor", "architecture"}),
		cd(agri, "Environmental Scientist", "Study environmental problems and develop solutions for pollution, conservation, and sustainability.", []string{"Environmental monitoring", "Data analysis", "Lab techniques", "Research methods", "Report writing", "Regulation knowledge", "Field work"}, 30000, 100000, 4, 78, []string{"environment", "science", "research", "conservation", "high-growth"}),
		cd(agri, "Veterinary Assistant", "Help veterinarians with animal care, treatments, and clinic operations.", []string{"Animal handling", "Clinic procedures", "Customer service", "Record keeping", "Sterilization", "Basic medical knowledge", "Compassion"}, 15000, 40000, 2, 70, []string{"agriculture", "animals", "healthcare", "hands-on", "entry-level"}),

		// === MANUFACTURING & INDUSTRY (15+) ===
		cd(manufacturing, "Factory Supervisor", "Oversee production lines, manage workers, and ensure quality control in manufacturing.", []string{"Production management", "Quality control", "Team leadership", "Safety compliance", "Scheduling", "Problem-solving", "Reporting"}, 25000, 70000, 3, 62, []string{"manufacturing", "management", "stable", "hands-on", "leadership"}),
		cd(manufacturing, "Quality Control Inspector", "Inspect products and materials to ensure they meet quality standards.", []string{"Inspection techniques", "Measurement tools", "Attention to detail", "Quality standards", "Report writing", "Process improvement", "Communication"}, 18000, 50000, 2, 60, []string{"manufacturing", "quality", "hands-on", "stable", "technical"}),
		cd(manufacturing, "Industrial Engineer", "Optimize production processes to improve efficiency, quality, and safety.", []string{"Process optimization", "Lean manufacturing", "Data analysis", "CAD software", "Project management", "Problem-solving", "Systems thinking"}, 35000, 120000, 4, 72, []string{"manufacturing", "engineering", "tech", "stable", "high-salary"}),
		cd(manufacturing, "Textile Worker", "Operate machinery to produce fabric and textile products in manufacturing settings.", []string{"Machine operation", "Fabric knowledge", "Quality checking", "Safety practices", "Attention to detail", "Physical stamina", "Teamwork"}, 12000, 35000, 2, 50, []string{"manufacturing", "textile", "hands-on", "entry-level", "stable"}),
		cd(manufacturing, "CNC Operator", "Operate computer-controlled machine tools to create precision parts.", []string{"CNC programming", "Machine setup", "Blueprint reading", "Precision measurement", "Tool selection", "Quality inspection", "Maintenance"}, 20000, 70000, 3, 68, []string{"manufacturing", "tech", "hands-on", "precision", "stable"}),

		// === TRANSPORTATION & LOGISTICS (15+) ===
		cd(transport, "Truck Driver", "Transport goods over long distances, ensuring timely and safe delivery.", []string{"Driving skills", "Route planning", "Vehicle maintenance", "Time management", "Customer communication", "Safety compliance", "Documentation"}, 20000, 70000, 2, 55, []string{"transport", "logistics", "driving", "stable", "essential"}),
		cd(transport, "Flight Attendant", "Ensure passenger safety and comfort during flights.", []string{"Safety procedures", "Customer service", "First aid", "Communication", "Conflict resolution", "Multilingual ability", "Teamwork"}, 25000, 80000, 2, 50, []string{"aviation", "hospitality", "travel", "people", "customer-service"}),
		cd(transport, "Logistics Coordinator", "Coordinate freight movement, track shipments, and manage supply chain operations.", []string{"Logistics software", "Supply chain knowledge", "Customer service", "Problem-solving", "Organization", "Communication", "Data entry"}, 20000, 60000, 2, 68, []string{"logistics", "business", "stable", "office-job", "operations"}),
		cd(transport, "Warehouse Manager", "Oversee warehouse operations including inventory, shipping, and receiving.", []string{"Inventory management", "Team leadership", "WMS software", "Safety compliance", "Space optimization", "Scheduling", "Reporting"}, 25000, 80000, 3, 62, []string{"logistics", "management", "stable", "hands-on", "operations"}),
		cd(transport, "Delivery Driver", "Deliver packages, food, and goods to customers within local areas.", []string{"Driving", "Route navigation", "Customer service", "Time management", "Physical fitness", "Basic vehicle maintenance", "Smartphone apps"}, 12000, 40000, 1, 55, []string{"delivery", "transport", "entry-level", "flexible", "gig-economy"}),
		cd(transport, "Shipping Clerk", "Process shipments, prepare documentation, and coordinate with carriers.", []string{"Shipping procedures", "Documentation", "Computer skills", "Organization", "Customer service", "Attention to detail", "Basic math"}, 15000, 45000, 2, 55, []string{"logistics", "office-job", "stable", "entry-level", "operations"}),

		// === EDUCATION (20+) ===
		cd(edu, "Online Tutor", "Teach students remotely via video platforms in various subjects.", []string{"Subject expertise", "Patience", "Technology proficiency", "Communication", "Class management", "Adaptability", "Scheduling"}, 15000, 50000, 2, 68, []string{"education", "remote-work", "teaching", "freelance", "flexible"}),
		cd(edu, "School Counselor", "Guide students in academic, career, and personal development.", []string{"Counseling skills", "Career guidance", "Communication", "Empathy", "Problem-solving", "Student assessment", "Program planning"}, 25000, 70000, 3, 65, []string{"education", "counseling", "helping-people", "stable", "school"}),
		cd(edu, "Early Childhood Educator", "Teach and care for young children in preschool and daycare settings.", []string{"Child development", "Patience", "Creative activities", "Classroom management", "Parent communication", "Safety awareness", "Nurturing attitude"}, 15000, 40000, 2, 68, []string{"education", "children", "teaching", "helping-people", "stable"}),
		cd(edu, "Special Education Teacher", "Teach students with disabilities, adapting curriculum to individual needs.", []string{"Special ed methods", "Patience", "Adaptability", "IEP development", "Behavior management", "Communication", "Empathy"}, 20000, 60000, 4, 75, []string{"education", "special-needs", "teaching", "helping-people", "rewarding"}),
		cd(edu, "Corporate Trainer", "Develop and deliver training programs for employees in organizations.", []string{"Training design", "Presentation skills", "Subject expertise", "Assessment creation", "Facilitation", "Communication", "Program evaluation"}, 30000, 100000, 3, 68, []string{"education", "corporate", "training", "business", "high-growth"}),
		cd(edu, "Language Instructor", "Teach languages (English, Nepali, etc.) to students of all levels.", []string{"Language proficiency", "Teaching methods", "Cultural knowledge", "Patience", "Communication", "Lesson planning", "Assessment"}, 15000, 50000, 2, 60, []string{"education", "language", "teaching", "freelance", "remote-work"}),
		cd(edu, "Curriculum Developer", "Design educational curricula, learning materials, and assessment tools.", []string{"Instructional design", "Subject expertise", "Writing skills", "Education technology", "Assessment design", "Research", "Creativity"}, 30000, 90000, 3, 68, []string{"education", "design", "writing", "creative", "stable"}),

		// === GOVERNMENT & SOCIAL (20+) ===
		cd(govt, "Civil Service Officer", "Work in government departments implementing policies and serving citizens.", []string{"Policy knowledge", "Administration", "Communication", "Problem-solving", "Integrity", "Public service", "Report writing"}, 25000, 80000, 3, 72, []string{"government", "stable", "prestigious", "service", "secure"}),
		cd(govt, "Social Worker", "Help individuals and families access resources and support services.", []string{"Empathy", "Case management", "Crisis intervention", "Community knowledge", "Communication", "Advocacy", "Documentation"}, 18000, 50000, 3, 65, []string{"social-work", "helping-people", "community", "stable", "rewarding"}),
		cd(govt, "NGO Program Manager", "Manage development programs for non-governmental organizations.", []string{"Program management", "Budgeting", "Monitoring and evaluation", "Proposal writing", "Stakeholder engagement", "Team leadership", "Reporting"}, 30000, 100000, 4, 65, []string{"ngo", "development", "management", "helping-people", "international"}),
		cd(govt, "Community Health Worker", "Provide basic health education and services in local communities.", []string{"Health knowledge", "Community outreach", "Communication", "Basic medical skills", "Record keeping", "Empathy", "Cultural sensitivity"}, 12000, 35000, 2, 70, []string{"healthcare", "community", "helping-people", "entry-level", "social"}),
		cd(govt, "Police Officer", "Maintain public safety, enforce laws, and respond to emergencies.", []string{"Law enforcement", "Physical fitness", "Communication", "Conflict resolution", "Integrity", "Community relations", "Emergency response"}, 20000, 60000, 4, 62, []string{"government", "security", "service", "stable", "discipline"}),
		cd(govt, "Firefighter", "Respond to fires, emergencies, and rescue situations to protect lives and property.", []string{"Fire suppression", "Emergency medical", "Physical fitness", "Teamwork", "Equipment operation", "Quick decision-making", "Stress management"}, 20000, 60000, 4, 68, []string{"government", "emergency", "service", "hands-on", "heroic"}),
		cd(govt, "Urban Planner", "Design and plan urban areas, managing land use and infrastructure development.", []string{"Planning software", "Land use knowledge", "Research", "Public consultation", "Regulation knowledge", "Design skills", "Project management"}, 30000, 100000, 4, 72, []string{"government", "design", "urban", "stable", "architecture"}),
		cd(govt, "Diplomat", "Represent Nepal internationally, handling foreign relations and diplomacy.", []string{"International relations", "Negotiation", "Language skills", "Cultural awareness", "Protocol knowledge", "Writing", "Analytical thinking"}, 40000, 150000, 5, 68, []string{"government", "international", "prestigious", "travel", "service"}),

		// === LEGAL (10+) ===
		cd(legalCat, "Paralegal", "Assist lawyers with legal research, document preparation, and case management.", []string{"Legal research", "Document drafting", "Case management", "Communication", "Organization", "Legal terminology", "Attention to detail"}, 20000, 60000, 3, 65, []string{"legal", "office-job", "stable", "research", "professional"}),
		cd(legalCat, "Legal Advisor", "Provide legal advice to organizations on compliance, contracts, and regulations.", []string{"Legal expertise", "Contract law", "Compliance", "Negotiation", "Risk assessment", "Communication", "Analytical thinking"}, 40000, 150000, 4, 72, []string{"legal", "corporate", "high-salary", "professional", "stable"}),
		cd(legalCat, "Notary Public", "Witness document signings and verify identities for legal transactions.", []string{"Legal knowledge", "Attention to detail", "Customer service", "Integrity", "Organization", "Document verification", "Business management"}, 15000, 50000, 2, 58, []string{"legal", "service", "office-job", "stable", "professional"}),
		cd(legalCat, "Human Rights Advocate", "Promote and protect human rights through advocacy, research, and legal action.", []string{"Human rights law", "Advocacy", "Research", "Communication", "Empathy", "Campaign skills", "Networking"}, 20000, 70000, 3, 70, []string{"legal", "social-justice", "ngo", "helping-people", "advocacy"}),
		cd(legalCat, "Corporate Lawyer", "Advise businesses on legal matters including contracts, mergers, and compliance.", []string{"Corporate law", "Contract drafting", "Negotiation", "M&A knowledge", "Client management", "Legal research", "Communication"}, 50000, 250000, 5, 72, []string{"legal", "corporate", "high-salary", "business", "prestigious"}),

		// === REAL ESTATE & CONSTRUCTION (15+) ===
		cd(realestate, "Real Estate Agent", "Help clients buy, sell, and rent properties, guiding them through transactions.", []string{"Sales skills", "Property knowledge", "Negotiation", "Marketing", "Customer service", "Local market knowledge", "Communication"}, 15000, 150000, 2, 55, []string{"real-estate", "sales", "freelance", "people", "flexible"}),
		cd(realestate, "Construction Manager", "Oversee construction projects, managing budgets, schedules, and teams.", []string{"Project management", "Budgeting", "Team leadership", "Safety compliance", "Blueprint reading", "Scheduling", "Client communication"}, 40000, 150000, 4, 68, []string{"construction", "management", "high-salary", "leadership", "stable"}),
		cd(realestate, "Architect", "Design buildings and structures, balancing aesthetics, function, and safety.", []string{"Design software (AutoCAD, Revit)", "Building codes", "Spatial planning", "Creativity", "Project management", "Client communication", "Technical knowledge"}, 30000, 150000, 4, 65, []string{"design", "construction", "creative", "high-salary", "professional"}),
		cd(realestate, "Surveyor", "Measure and map land areas for construction, property boundaries, and development.", []string{"Surveying equipment", "GPS technology", "Mapping software", "Mathematics", "Field work", "Attention to detail", "Report writing"}, 25000, 80000, 3, 65, []string{"construction", "land", "outdoor", "technical", "stable"}),
		cd(realestate, "Property Developer", "Acquire land and develop properties for residential or commercial use.", []string{"Project feasibility", "Market analysis", "Financial planning", "Stakeholder management", "Negotiation", "Risk assessment", "Team coordination"}, 50000, 500000, 5, 62, []string{"real-estate", "business", "entrepreneur", "high-risk", "high-salary"}),
		cd(realestate, "Interior Decorator", "Decorate interior spaces with furniture, colors, and accessories.", []string{"Design taste", "Color coordination", "Product sourcing", "Client consultation", "Budget management", "Trend awareness", "Communication"}, 20000, 70000, 2, 58, []string{"design", "creative", "real-estate", "freelance", "art"}),

		// === BEAUTY & WELLNESS (15+) ===
		cd(beauty, "Makeup Artist", "Apply makeup for events, photoshoots, film, and television productions.", []string{"Makeup techniques", "Product knowledge", "Color matching", "Skin care basics", "Client consultation", "Hygiene practices", "Portfolio building"}, 15000, 70000, 2, 50, []string{"beauty", "creative", "freelance", "events", "art"}),
		cd(beauty, "Hairstylist", "Cut, color, and style hair for clients in salons or freelance settings.", []string{"Hair cutting", "Coloring techniques", "Styling", "Client consultation", "Product knowledge", "Sanitation", "Trend awareness"}, 15000, 50000, 2, 52, []string{"beauty", "creative", "hands-on", "freelance", "stable"}),
		cd(beauty, "Spa Therapist", "Provide massage, skincare, and wellness treatments to clients.", []string{"Massage techniques", "Skin care", "Aromatherapy", "Client care", "Hygiene", "Product knowledge", "Communication"}, 15000, 50000, 2, 60, []string{"wellness", "beauty", "hands-on", "helping-people", "stable"}),
		cd(beauty, "Nail Technician", "Provide manicure, pedicure, and nail art services to clients.", []string{"Nail care", "Nail art", "Sanitation", "Product knowledge", "Customer service", "Creativity", "Attention to detail"}, 12000, 40000, 2, 48, []string{"beauty", "creative", "hands-on", "freelance", "stable"}),
		cd(beauty, "Aesthetician", "Provide skincare treatments like facials, peels, and microdermabrasion.", []string{"Skin analysis", "Facial treatments", "Product knowledge", "Client consultation", "Sanitation", "Trend awareness", "Communication"}, 18000, 60000, 2, 60, []string{"beauty", "wellness", "healthcare", "hands-on", "stable"}),

		// === SCIENCE & RESEARCH (15+) ===
		cd(science, "Research Scientist", "Conduct scientific research to advance knowledge in various fields.", []string{"Research methods", "Data analysis", "Lab techniques", "Scientific writing", "Critical thinking", "Project management", "Grant writing"}, 30000, 120000, 5, 72, []string{"science", "research", "academic", "high-salary", "prestigious"}),
		cd(science, "Chemist", "Study chemical substances and develop products for pharmaceuticals, manufacturing, and research.", []string{"Lab techniques", "Chemical analysis", "Safety protocols", "Data interpretation", "Report writing", "Problem-solving", "Equipment operation"}, 25000, 90000, 4, 68, []string{"science", "chemistry", "lab", "research", "stable"}),
		cd(science, "Biologist", "Study living organisms and their environments for research and conservation.", []string{"Lab techniques", "Field research", "Data analysis", "Microscopy", "Report writing", "Critical thinking", "Environmental knowledge"}, 25000, 80000, 4, 70, []string{"science", "biology", "research", "environment", "lab"}),
		cd(science, "Geologist", "Study the Earth's physical structure and materials for resource exploration and environmental assessment.", []string{"Geological mapping", "Field surveys", "Lab analysis", "GIS software", "Report writing", "Mineral identification", "Safety awareness"}, 30000, 100000, 4, 72, []string{"science", "geology", "outdoor", "research", "stable"}),
		cd(science, "Meteorologist", "Study weather patterns and climate to forecast conditions and understand climate change.", []string{"Weather modeling", "Data analysis", "Satellite imagery", "Statistics", "Communication", "Computer programming", "Research methods"}, 25000, 90000, 4, 72, []string{"science", "weather", "research", "tech", "stable"}),
		cd(science, "Astronomer", "Study celestial objects and phenomena to advance understanding of the universe.", []string{"Telescope operation", "Data analysis", "Physics knowledge", "Mathematics", "Computer modeling", "Research methods", "Scientific writing"}, 30000, 120000, 5, 78, []string{"science", "space", "research", "academic", "prestigious"}),

		// === VOCATIONAL & TECHNICAL (20+) ===
		cd(vocation, "Electric Vehicle Technician", "Specialize in repairing and maintaining electric vehicles and charging infrastructure.", []string{"EV systems", "High-voltage safety", "Diagnostic software", "Battery technology", "Electrical engineering", "Customer service", "Problem-solving"}, 30000, 100000, 3, 88, []string{"automotive", "green", "tech", "hands-on", "high-growth"}),
		cd(vocation, "Computer Repair Technician", "Diagnose and repair computer hardware and software issues.", []string{"Hardware diagnostics", "OS knowledge", "Component replacement", "Troubleshooting", "Customer service", "Network basics", "Data recovery"}, 15000, 50000, 2, 55, []string{"tech", "repair", "hands-on", "stable", "entry-level"}),
		cd(vocation, "Mobile Phone Repair Technician", "Repair smartphones and tablets, replacing screens, batteries, and components.", []string{"Micro-soldering", "Component replacement", "Diagnostic tools", "Customer service", "Fine motor skills", "Parts knowledge", "Business management"}, 15000, 50000, 2, 55, []string{"tech", "repair", "hands-on", "entry-level", "business"}),
		cd(vocation, "Plumber", "Install and repair water, gas, and drainage systems in buildings.", []string{"Pipe fitting", "Blueprint reading", "Tools expertise", "Troubleshooting", "Customer service", "Safety compliance", "Physical stamina"}, 20000, 70000, 3, 85, []string{"trades", "hands-on", "essential", "stable", "no-college"}),
		cd(vocation, "Baker", "Prepare and bake bread, pastries, and other baked goods.", []string{"Baking techniques", "Recipe following", "Creativity", "Time management", "Sanitation", "Precision", "Product presentation"}, 12000, 40000, 2, 52, []string{"food", "hospitality", "creative", "hands-on", "stable"}),
		cd(vocation, "Tailor", "Alter and create garments, providing custom fitting and repair services.", []string{"Sewing", "Pattern making", "Fabric knowledge", "Measuring", "Customer consultation", "Attention to detail", "Business skills"}, 12000, 40000, 2, 50, []string{"fashion", "hands-on", "creative", "business", "stable"}),
		cd(vocation, "Chef", "Prepare and cook food in restaurants, hotels, and other food service establishments.", []string{"Cooking techniques", "Menu planning", "Kitchen management", "Creativity", "Time management", "Sanitation", "Team leadership"}, 20000, 80000, 3, 52, []string{"hospitality", "food", "creative", "hands-on", "stable"}),

		// === E-COMMERCE & RETAIL (15+) ===
		cd(ecommerce, "Retail Store Manager", "Oversee retail store operations, manage staff, and drive sales.", []string{"Sales management", "Inventory control", "Staff supervision", "Customer service", "Visual merchandising", "Financial reporting", "Scheduling"}, 20000, 70000, 3, 55, []string{"retail", "management", "business", "people", "stable"}),
		cd(ecommerce, "Visual Merchandiser", "Design product displays and store layouts to maximize sales.", []string{"Design sense", "Product knowledge", "Creativity", "Trend awareness", "Attention to detail", "Communication", "Budget management"}, 18000, 50000, 2, 55, []string{"retail", "design", "creative", "business", "marketing"}),
		cd(ecommerce, "Dropshipping Entrepreneur", "Run an e-commerce business without holding inventory, shipping directly from suppliers.", []string{"E-commerce platforms", "Product research", "Marketing", "Customer service", "Supplier management", "Analytics", "Advertising"}, 10000, 200000, 3, 55, []string{"ecommerce", "entrepreneur", "remote-work", "digital", "freelance"}),
		cd(ecommerce, "Amazon FBA Seller", "Sell products on Amazon using Fulfilled by Amazon service for storage and shipping.", []string{"Product sourcing", "Listing optimization", "PPC advertising", "Inventory management", "Supplier relations", "Customer service", "Analytics"}, 15000, 200000, 3, 58, []string{"ecommerce", "entrepreneur", "business", "digital", "freelance"}),
		cd(ecommerce, "Customer Service Representative", "Handle customer inquiries, complaints, and support for businesses.", []string{"Communication", "Problem-solving", "Product knowledge", "Patience", "CRM software", "Multitasking", "Empathy"}, 12000, 35000, 2, 55, []string{"customer-service", "entry-level", "stable", "office-job", "remote-work"}),
		cd(ecommerce, "Cashier", "Process customer transactions at retail checkout counters.", []string{"Cash handling", "Customer service", "Basic math", "POS systems", "Communication", "Speed and accuracy", "Teamwork"}, 10000, 25000, 1, 40, []string{"retail", "entry-level", "customer-service", "stable", "part-time"}),

		// === MORE EMERGING DIGITAL CAREERS ===
		cd(digital, "Digital Content Strategist", "Plan and execute content strategies for brands across digital platforms.", []string{"Content strategy", "SEO", "Analytics", "Editorial planning", "Brand voice", "Audience research", "Team coordination"}, 20000, 60000, 3, 70, []string{"content", "strategy", "digital", "marketing", "remote-work"}),
		cd(digital, "Social Media Analyst", "Analyze social media performance data to optimize content and engagement strategies.", []string{"Analytics tools", "Data interpretation", "Platform expertise", "Reporting", "Trend analysis", "Excel", "Communication"}, 18000, 50000, 2, 68, []string{"analytics", "social-media", "digital", "data", "marketing"}),
		cd(digital, "Email Marketing Specialist", "Design and manage email marketing campaigns to drive customer engagement and sales.", []string{"Email platforms (Mailchimp)", "Copywriting", "List management", "A/B testing", "Analytics", "Segmentation", "Automation"}, 18000, 50000, 2, 65, []string{"marketing", "email", "digital", "remote-work", "freelance"}),
		cd(digital, "Affiliate Marketing Manager", "Manage affiliate programs and partnerships to drive revenue through referral traffic.", []string{"Affiliate platforms", "Partner recruitment", "Commission management", "Analytics", "Relationship management", "Marketing", "Negotiation"}, 20000, 70000, 3, 62, []string{"marketing", "digital", "business", "freelance", "remote-work"}),
		cd(digital, "Product Manager", "Define product vision, strategy, and roadmap for digital products.", []string{"Product strategy", "User research", "Roadmap planning", "Stakeholder management", "Data analysis", "Agile methodology", "Communication"}, 40000, 120000, 4, 78, []string{"tech", "product", "management", "high-growth", "leadership"}),
		cd(digital, "Scrum Master", "Facilitate Agile development processes and help teams deliver effectively.", []string{"Agile/Scrum expertise", "Facilitation", "Coaching", "Conflict resolution", "Project tracking tools", "Communication", "Team empowerment"}, 35000, 100000, 3, 72, []string{"tech", "agile", "management", "leadership", "stable"}),
		cd(digital, "Site Reliability Engineer", "Ensure systems are reliable, scalable, and performant through automation and monitoring.", []string{"Linux", "Cloud platforms", "Automation", "Monitoring tools", "Incident response", "Programming", "Documentation"}, 45000, 150000, 4, 90, []string{"tech", "devops", "high-salary", "remote-work", "engineering"}),
		cd(digital, "Database Administrator", "Manage and maintain database systems ensuring performance, security, and availability.", []string{"SQL", "Database management", "Backup/recovery", "Performance tuning", "Security", "Scripting", "Problem-solving"}, 30000, 100000, 4, 72, []string{"tech", "database", "stable", "enterprise", "engineering"}),
		cd(digital, "Network Engineer", "Design, implement, and manage computer networks for organizations.", []string{"Network protocols", "Router/switch config", "Security", "Troubleshooting", "Firewall management", "VPN", "Documentation"}, 25000, 90000, 3, 70, []string{"tech", "networking", "stable", "enterprise", "engineering"}),
		cd(digital, "Systems Administrator", "Manage and maintain an organization's IT infrastructure and servers.", []string{"OS administration", "Server management", "Security", "Backup/recovery", "Scripting", "Troubleshooting", "Documentation"}, 20000, 70000, 3, 65, []string{"tech", "it", "stable", "enterprise", "hands-on"}),
		cd(digital, "IT Support Specialist", "Provide technical support to users, resolving hardware and software issues.", []string{"Troubleshooting", "Customer service", "OS knowledge", "Network basics", "Hardware knowledge", "Documentation", "Patience"}, 15000, 45000, 2, 55, []string{"tech", "support", "entry-level", "stable", "helping-people"}),
		cd(digital, "API Developer", "Design and build APIs that connect applications and services.", []string{"REST/GraphQL", "Programming (Node/Python)", "API security", "Documentation", "Testing", "Microservices", "Problem-solving"}, 35000, 120000, 3, 82, []string{"tech", "api", "remote-work", "engineering", "high-growth"}),
		cd(digital, "Low-Code Developer", "Build applications using low-code platforms with minimal hand-coding.", []string{"Low-code platforms", "Process automation", "Integration", "Problem-solving", "Logic design", "Testing", "Deployment"}, 20000, 70000, 2, 72, []string{"tech", "low-code", "emerging", "entry-level", "engineering"}),
		cd(digital, "No-Code Entrepreneur", "Build and launch digital businesses using no-code tools without programming.", []string{"No-code platforms", "Product thinking", "Marketing", "Customer research", "Automation", "Design thinking", "Business strategy"}, 10000, 80000, 2, 65, []string{"entrepreneur", "no-code", "digital", "creative", "freelance"}),
		cd(digital, "Digital Nomad", "Work remotely while traveling, combining location independence with online work.", []string{"Remote work discipline", "Time management", "Self-motivation", "Digital skills", "Budgeting", "Cultural adaptability", "Communication"}, 15000, 100000, 3, 60, []string{"remote-work", "freelance", "travel", "digital", "lifestyle"}),

		// === MORE CREATIVE ARTS ===
		cd(creative, "Set Designer", "Design and create sets for theater, film, television, and events.", []string{"Design software", "Construction knowledge", "Creativity", "Project management", "Collaboration", "Budget management", "Spatial awareness"}, 20000, 60000, 3, 48, []string{"design", "creative", "entertainment", "hands-on", "events"}),
		cd(creative, "Art Director", "Oversee visual style and creative direction for media, advertising, and design projects.", []string{"Creative direction", "Team leadership", "Visual design", "Client presentation", "Budget management", "Trend awareness", "Communication"}, 35000, 100000, 4, 60, []string{"design", "creative", "leadership", "media", "high-salary"}),
		cd(creative, "Brand Identity Designer", "Create visual brand identities including logos, color palettes, and brand guidelines.", []string{"Brand strategy", "Logo design", "Visual systems", "Typography", "Color theory", "Client presentation", "Market research"}, 25000, 80000, 3, 62, []string{"design", "branding", "creative", "freelance", "business"}),
		cd(creative, "Packaging Designer", "Design product packaging that attracts consumers and communicates brand values.", []string{"3D design", "Structural packaging", "Print production", "Material knowledge", "Brand alignment", "Creativity", "Consumer psychology"}, 20000, 65000, 3, 55, []string{"design", "creative", "product", "manufacturing", "freelance"}),
		cd(creative, "Textile Designer", "Create designs and patterns for fabric and textile products.", []string{"Pattern design", "Color theory", "Fabric knowledge", "Digital design tools", "Trend research", "Production knowledge", "Creativity"}, 18000, 50000, 3, 55, []string{"design", "textile", "creative", "fashion", "manufacturing"}),

		// === MORE HEALTHCARE ===
		cd(health, "Medical Transcriptionist", "Convert voice recordings from healthcare professionals into written reports.", []string{"Typing speed", "Medical terminology", "Listening skills", "Grammar", "Attention to detail", "Confidentiality", "Software proficiency"}, 12000, 35000, 2, 50, []string{"healthcare", "office-job", "remote-work", "entry-level", "stable"}),
		cd(health, "Health Educator", "Teach communities about healthy behaviors and disease prevention.", []string{"Health knowledge", "Teaching", "Communication", "Program planning", "Community outreach", "Cultural sensitivity", "Evaluation"}, 18000, 50000, 3, 65, []string{"healthcare", "education", "community", "helping-people", "stable"}),
		cd(health, "Medical Coder", "Assign standardized codes to medical diagnoses and procedures for billing.", []string{"Medical coding systems", "Anatomy knowledge", "Attention to detail", "Software proficiency", "Confidentiality", "Regulation knowledge", "Analytical skills"}, 15000, 45000, 3, 68, []string{"healthcare", "office-job", "remote-work", "stable", "entry-level"}),
		cd(health, "Clinical Research Coordinator", "Manage clinical trials and research studies in healthcare settings.", []string{"Research methodology", "Regulation compliance", "Data management", "Patient coordination", "Documentation", "Communication", "Organization"}, 25000, 70000, 4, 72, []string{"healthcare", "research", "clinical", "stable", "professional"}),
		cd(health, "Occupational Therapist", "Help patients develop or recover daily living skills after injury or illness.", []string{"Therapy techniques", "Patient assessment", "Treatment planning", "Adaptive equipment", "Patient education", "Documentation", "Empathy"}, 25000, 80000, 4, 78, []string{"healthcare", "therapy", "helping-people", "clinical", "stable"}),
		cd(health, "Speech Therapist", "Diagnose and treat speech, language, and communication disorders.", []string{"Speech assessment", "Therapy techniques", "Patient education", "Documentation", "Communication", "Patience", "Creativity"}, 25000, 70000, 4, 75, []string{"healthcare", "therapy", "helping-people", "clinical", "stable"}),

		// === MORE AGRICULTURE ===
		cd(agri, "Tea Garden Manager", "Manage tea plantation operations from cultivation to processing and sales.", []string{"Tea cultivation", "Garden management", "Processing knowledge", "Staff supervision", "Quality control", "Marketing", "Sustainability"}, 20000, 60000, 3, 60, []string{"agriculture", "management", "nepal-specific", "outdoor", "business"}),
		cd(agri, "Poultry Farmer", "Raise chickens and other poultry for meat and egg production.", []string{"Animal husbandry", "Disease management", "Feed formulation", "Biosecurity", "Record keeping", "Business management", "Marketing"}, 15000, 50000, 3, 62, []string{"agriculture", "farming", "hands-on", "business", "nepal-specific"}),
		cd(agri, "Mushroom Farmer", "Cultivate mushrooms for local markets and commercial sale.", []string{"Mushroom cultivation", "Climate control", "Sterilization", "Harvesting", "Marketing", "Business planning", "Quality control"}, 12000, 40000, 2, 65, []string{"agriculture", "farming", "entrepreneur", "nepal-specific", "hands-on"}),
		cd(agri, "Beekeeper", "Maintain bee colonies for honey production and pollination services.", []string{"Bee colony management", "Honey extraction", "Queen rearing", "Disease control", "Equipment maintenance", "Marketing", "Seasonal planning"}, 12000, 40000, 2, 65, []string{"agriculture", "beekeeping", "outdoor", "entrepreneur", "nepal-specific"}),
		cd(agri, "Hydroponics Farmer", "Grow plants without soil using nutrient-rich water solutions in controlled environments.", []string{"Hydroponic systems", "Nutrient management", "Climate control", "Plant science", "Problem-solving", "Business planning", "Marketing"}, 18000, 60000, 3, 72, []string{"agriculture", "technology", "emerging", "entrepreneur", "sustainable"}),

		// === MORE EDUCATION ===
		cd(edu, "Montessori Teacher", "Guide children through Montessori educational methods in early childhood settings.", []string{"Montessori method", "Child development", "Observation", "Classroom preparation", "Patience", "Creativity", "Parent communication"}, 12000, 35000, 2, 65, []string{"education", "children", "teaching", "alternative", "stable"}),
		cd(edu, "University Professor", "Teach and conduct research at colleges and universities.", []string{"Subject expertise", "Research", "Teaching", "Mentoring", "Publishing", "Grant writing", "Curriculum development"}, 35000, 100000, 5, 68, []string{"education", "academic", "research", "prestigious", "stable"}),
		cd(edu, "Education Technology Specialist", "Integrate technology into educational settings to enhance learning outcomes.", []string{"EdTech tools", "Training", "Technical support", "Curriculum integration", "Project management", "Communication", "Problem-solving"}, 25000, 70000, 3, 75, []string{"education", "technology", "edtech", "high-growth", "remote-work"}),
		cd(edu, "English Language Teacher", "Teach English to non-native speakers in schools, language institutes, or online.", []string{"Language teaching", "Lesson planning", "Patience", "Cross-cultural communication", "Classroom management", "Assessment", "Technology proficiency"}, 15000, 50000, 2, 60, []string{"education", "language", "teaching", "remote-work", "stable"}),
		cd(edu, "Mathematics Tutor", "Provide one-on-one or small group mathematics instruction to students.", []string{"Math expertise", "Patience", "Communication", "Adaptability", "Assessment", "Technology tools", "Encouragement"}, 12000, 40000, 2, 60, []string{"education", "math", "tutoring", "freelance", "flexible"}),

		// === MORE TOURISM ===
		cd(tourism, "Trekking Guide", "Lead trekking expeditions in Nepal's mountains, ensuring safety and memorable experiences.", []string{"Mountain navigation", "First aid", "Language skills", "Cultural knowledge", "Physical fitness", "Group management", "Emergency response"}, 15000, 50000, 3, 55, []string{"tourism", "outdoor", "adventure", "nepal-specific", "hands-on"}),
		cd(tourism, "Ayurvedic Therapist", "Provide traditional Ayurvedic treatments and wellness therapies to clients.", []string{"Ayurveda knowledge", "Massage techniques", "Herbal medicine", "Client consultation", "Wellness counseling", "Sanitation", "Communication"}, 15000, 45000, 2, 60, []string{"wellness", "traditional", "healthcare", "nepal-specific", "tourism"}),
		cd(tourism, "Eco-Lodge Operator", "Run environmentally sustainable accommodations for eco-tourists.", []string{"Hospitality management", "Sustainability practices", "Customer service", "Local community relations", "Marketing", "Financial management", "Environmental knowledge"}, 20000, 70000, 3, 65, []string{"tourism", "sustainable", "hospitality", "entrepreneur", "nepal-specific"}),

		// === MORE MANUFACTURING ===
		cd(manufacturing, "Pashmina Craftsperson", "Create high-quality pashmina shawls and garments using traditional techniques.", []string{"Weaving", "Fiber knowledge", "Quality control", "Design sense", "Traditional techniques", "Patience", "Attention to detail"}, 12000, 40000, 3, 50, []string{"handicraft", "traditional", "nepal-specific", "creative", "export"}),
		cd(manufacturing, "Carpet Weaver", "Create hand-knotted carpets using traditional Nepali weaving techniques.", []string{"Weaving techniques", "Pattern following", "Color coordination", "Quality control", "Patience", "Physical stamina", "Attention to detail"}, 12000, 35000, 3, 48, []string{"handicraft", "traditional", "nepal-specific", "hands-on", "export"}),
		cd(manufacturing, "Food Processing Technician", "Operate equipment to process and package food products.", []string{"Food safety", "Equipment operation", "Quality control", "Sanitation", "Production tracking", "Maintenance", "Teamwork"}, 15000, 40000, 2, 60, []string{"manufacturing", "food", "hands-on", "stable", "entry-level"}),
		cd(manufacturing, "Garment Factory Worker", "Operate sewing machines and production equipment in garment manufacturing.", []string{"Sewing", "Machine operation", "Quality checking", "Speed and accuracy", "Teamwork", "Safety awareness", "Attention to detail"}, 10000, 25000, 2, 45, []string{"manufacturing", "garment", "hands-on", "entry-level", "stable"}),

		// === MORE SOCIAL & GOVERNMENT ===
		cd(govt, "Local Government Officer", "Implement development programs and deliver services at the municipal level.", []string{"Administration", "Program implementation", "Community engagement", "Report writing", "Budget management", "Regulation knowledge", "Communication"}, 20000, 55000, 3, 70, []string{"government", "local", "service", "stable", "nepal-specific"}),
		cd(govt, "Tax Officer", "Assess and collect taxes, ensuring compliance with tax laws.", []string{"Tax law knowledge", "Assessment", "Audit skills", "Mathematics", "Attention to detail", "Integrity", "Communication"}, 25000, 70000, 3, 68, []string{"government", "finance", "stable", "professional", "nepal-specific"}),
		cd(govt, "Foreign Employment Counselor", "Advise and assist Nepali citizens seeking employment opportunities abroad.", []string{"Labor market knowledge", "Counseling", "Documentation", "Regulation knowledge", "Communication", "Language skills", "Ethical practice"}, 18000, 50000, 3, 55, []string{"government", "employment", "service", "nepal-specific", "people"}),
		cd(govt, "Red Cross Worker", "Provide humanitarian aid, disaster response, and community health services.", []string{"First aid", "Disaster response", "Community outreach", "Supply management", "Communication", "Empathy", "Team coordination"}, 15000, 45000, 3, 65, []string{"ngo", "humanitarian", "helping-people", "stable", "service"}),

		// === MORE FINANCE ===
		cd(biz, "Microfinance Officer", "Provide small loans and financial services to low-income individuals and communities.", []string{"Microfinance principles", "Client assessment", "Portfolio management", "Community outreach", "Repayment tracking", "Communication", "Financial literacy"}, 18000, 50000, 3, 62, []string{"finance", "microfinance", "nepal-specific", "helping-people", "community"}),
		cd(biz, "Cooperative Manager", "Manage the operations of cooperative societies and ensure member benefits.", []string{"Cooperative principles", "Financial management", "Member services", "Regulation compliance", "Leadership", "Community relations", "Reporting"}, 20000, 55000, 3, 60, []string{"finance", "cooperative", "nepal-specific", "management", "community"}),
		cd(biz, "Insurance Agent", "Sell insurance policies and advise clients on risk management.", []string{"Insurance products", "Sales", "Client relationship", "Risk assessment", "Documentation", "Communication", "Persuasion"}, 15000, 60000, 2, 58, []string{"finance", "insurance", "sales", "freelance", "people"}),
		cd(biz, "Remittance Agent", "Facilitate money transfers for Nepali workers abroad sending money home.", []string{"Financial transactions", "Currency exchange", "Customer service", "Documentation", "Regulation knowledge", "Record keeping", "Communication"}, 12000, 35000, 2, 52, []string{"finance", "remittance", "nepal-specific", "service", "entry-level"}),

		// === MORE HEALTHCARE SUPPORT ===
		cd(health, "Dental Assistant", "Assist dentists during procedures and manage clinic operations.", []string{"Chairside assistance", "Sterilization", "X-ray operation", "Appointment management", "Patient care", "Inventory management", "Communication"}, 12000, 35000, 2, 68, []string{"healthcare", "dental", "clinical", "entry-level", "stable"}),
		cd(health, "Pharmacy Technician", "Assist pharmacists in dispensing medications and managing inventory.", []string{"Medication knowledge", "Inventory management", "Customer service", "Prescription processing", "Labeling", "Regulation compliance", "Attention to detail"}, 12000, 35000, 2, 65, []string{"healthcare", "pharmacy", "clinical", "entry-level", "stable"}),
		cd(health, "Home Health Aide", "Provide in-home care and assistance to elderly, disabled, or recovering patients.", []string{"Patient care", "Personal hygiene assistance", "Meal preparation", "Medication reminders", "Companionship", "Observation", "Communication"}, 10000, 30000, 2, 65, []string{"healthcare", "elderly-care", "helping-people", "entry-level", "stable"}),

		// === MORE MEDIA ===
		cd(media, "Photojournalist", "Capture news-worthy photographs for media publications and digital platforms.", []string{"Photography", "News judgment", "Quick reaction", "Editing software", "Storytelling", "Ethics", "Fieldwork"}, 15000, 50000, 3, 50, []string{"media", "photography", "journalism", "creative", "fieldwork"}),
		cd(media, "Documentary Filmmaker", "Research, film, and produce documentary content on social and cultural topics.", []string{"Filmmaking", "Research", "Interviewing", "Editing", "Storytelling", "Project management", "Distribution"}, 20000, 80000, 4, 58, []string{"media", "film", "creative", "storytelling", "freelance"}),
		cd(media, "Subtitle Creator", "Create and synchronize subtitles for films, videos, and online content.", []string{"Language skills", "Typing speed", "Timing accuracy", "Translation", "Software proficiency", "Attention to detail", "Formatting"}, 10000, 30000, 2, 55, []string{"media", "language", "freelance", "remote-work", "entry-level"}),
		cd(media, "Audiobook Narrator", "Record and narrate audiobooks with proper pacing and character voices.", []string{"Voice control", "Recording setup", "Script interpretation", "Character voices", "Pacing", "Audio editing basics", "Stamina"}, 15000, 50000, 2, 55, []string{"media", "voice", "creative", "freelance", "remote-work"}),

		cd(digital, "NFT Artist", "Create and sell digital artwork as non-fungible tokens on blockchain platforms.", []string{"Digital art creation", "Blockchain basics", "Marketing", "Community building", "Creativity", "Social media", "Trend awareness"}, 15000, 200000, 3, 45, []string{"art", "digital", "blockchain", "creative", "freelance"}),
		cd(digital, "Metaverse Developer", "Build virtual worlds and experiences in the metaverse.", []string{"3D development", "Unity/Unreal", "Blockchain integration", "Game design", "Creativity", "Full-stack skills", "Spatial design"}, 40000, 200000, 5, 72, []string{"tech", "metaverse", "creative", "emerging", "high-growth"}),
		cd(digital, "DevOps Engineer", "Bridge development and operations, automating infrastructure and deployment.", []string{"CI/CD pipelines", "Cloud platforms", "Containerization", "Infrastructure as Code", "Monitoring", "Scripting", "Linux"}, 50000, 200000, 4, 90, []string{"tech", "devops", "high-salary", "remote-work", "engineering"}),
		cd(digital, "Data Engineer", "Build and maintain data pipelines and infrastructure for analytics.", []string{"Python", "SQL", "ETL pipelines", "Data warehousing", "Big data tools", "Cloud platforms", "Data modeling"}, 45000, 180000, 4, 88, []string{"tech", "data", "engineering", "high-growth", "remote-work"}),
		cd(digital, "Full-Stack Developer", "Build both frontend and backend of web applications.", []string{"JavaScript/TypeScript", "React/Vue/Angular", "Node.js/Python", "Databases", "APIs", "Git", "Problem-solving"}, 35000, 150000, 3, 82, []string{"tech", "coding", "remote-work", "freelance", "high-growth"}),
		cd(digital, "Mobile App Developer", "Build applications for iOS and Android platforms.", []string{"Swift/Kotlin", "React Native/Flutter", "API integration", "UI implementation", "App store deployment", "Testing", "Problem-solving"}, 35000, 150000, 3, 78, []string{"tech", "mobile", "remote-work", "freelance", "high-growth"}),
		cd(digital, "Technical Writer", "Create documentation, tutorials, and technical content for software and products.", []string{"Technical writing", "Documentation tools", "Subject research", "Clarity and precision", "User empathy", "Information architecture", "Collaboration"}, 25000, 90000, 3, 72, []string{"writing", "tech", "remote-work", "freelance", "stable"}),
		cd(digital, "QA Tester", "Test software applications to find bugs and ensure quality before release.", []string{"Testing methodologies", "Bug reporting", "Test automation", "Attention to detail", "Problem-solving", "Communication", "Basic programming"}, 20000, 70000, 2, 68, []string{"tech", "testing", "entry-level", "remote-work", "stable"}),
	}

	// Convert definitions to careerSeed slice
	var result []careerSeed
	for _, d := range defs {
		tmpl := templates[d.cat.slug]
		if tmpl.tasks == nil {
			tmpl = templates["technology-it"]
		}
		if d.studyDur == "" {
			d.studyDur = "2-4 years"
		}
		if d.degree == "" {
			d.degree = tmpl.edu
		}
		outlook := d.cat.name + " sector in Nepal. " + tmpl.outlook
		slug := d.title
		// Build a simple description
		desc := d.title + " is a career in the " + d.cat.name + " industry. " + d.summary + " Skills needed include " + joinSkills(d.skills) + "."
		se := careerSeed{
			CategoryName:    d.cat.name,
			CategorySlug:    d.cat.slug,
			CategoryDesc:    d.cat.desc,
			CategoryIcon:    d.cat.icon,
			Title:           d.title,
			Slug:            generateSlug(slug),
			Summary:         d.summary,
			Description:     desc,
			DailyTasks:      tmpl.tasks,
			Skills:          d.skills,
			SalaryMin:       d.salMin,
			SalaryMax:       d.salMax,
			SalaryCurrency:  "NPR",
			SalaryPeriod:    "monthly",
			Difficulty:      d.diff,
			FutureProof:     d.future,
			EducationReq:    d.degree,
			Outlook:         outlook,
			Tags:            d.tags,
			WorkLifeBalance: d.workLife,
			StudyDuration:   d.studyDur,
			DegreeRequired:  d.degree,
			CreativeScore:   d.creative,
			TechnicalScore:  d.technical,
			FreelancePot:    d.freelance,
			IsGovernment:    d.govt,
			IsRemoteOK:      d.remote,
			SalaryTiersJSON: fmt.Sprintf(`{"entry":{"min":%d,"max":%d,"currency":"NPR","period":"monthly"},"average":{"min":%d,"max":%d,"currency":"NPR","period":"monthly"},"experienced":{"min":%d,"max":%d,"currency":"NPR","period":"monthly"},"freelance":{"min":%d,"max":%d,"currency":"NPR","period":"monthly"}}`,
				d.salMin*60/100, d.salMin, d.salMin, d.salMax, d.salMax, d.salMax*15/10, d.salMin, d.salMax*15/10),
			DemandDataJSON:  `{"trend":"growing","growth_forecast":"positive","opportunities":"moderate","global_opportunity":true,"nepal_demand":"growing"}`,
			SourceLabelsJSON: `{"last_updated":"2026-01","salary_source":"Carevo Market Research 2026","demand_source":"Nepal Labor Market Survey","is_verified":false,"confidence_score":65}`,
			MetadataJSON:    fmt.Sprintf(`{"ai_proof_score":%d,"freelance_potential":%d,"burnout_risk":%d,"remote_potential":%d}`, d.future, d.freelance, 6-d.workLife, boolToInt(d.remote)*5),
		}
		result = append(result, se)
	}
	return result
}

func generateSlug(title string) string {
	slug := ""
	for _, c := range title {
		if c >= 'a' && c <= 'z' {
			slug += string(c)
		} else if c >= 'A' && c <= 'Z' {
			slug += string(c + 32)
		} else if c == ' ' || c == '/' {
			slug += "-"
		}
	}
	return slug
}

func joinSkills(skills []string) string {
	if len(skills) == 0 {
		return ""
	}
	result := ""
	for i, s := range skills {
		if i > 0 && i == len(skills)-1 {
			result += " and "
		} else if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
