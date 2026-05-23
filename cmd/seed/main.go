package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/guruorgoru/carevo/internal/config"
	"github.com/guruorgoru/carevo/internal/database"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type careerSeed struct {
	CategoryName     string
	CategorySlug     string
	CategoryDesc     string
	CategoryIcon     string
	Title            string
	Slug             string
	Summary          string
	Description      string
	DailyTasks       []string
	Skills           []string
	SalaryMin        int
	SalaryMax        int
	SalaryCurrency   string
	SalaryPeriod     string
	Difficulty       int
	FutureProof      int
	EducationReq     string
	Outlook          string
	Tags             []string
	Resources        []resourceSeed
	RoadmapSteps     []roadmapSeed
	WorkLifeBalance  int
	StudyDuration    string
	DegreeRequired   string
	CreativeScore    int
	TechnicalScore   int
	FreelancePot     int
	IsGovernment     bool
	IsRemoteOK       bool
	ExamRequired     string
	SalaryTiersJSON  string
	CitySalariesJSON string
	DemandDataJSON   string
	SourceLabelsJSON string
	NepaliContentJSON string
	MetadataJSON     string
}

type resourceSeed struct {
	Title       string
	URL         string
	Description string
}

type roadmapLink struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type roadmapSeed struct {
	StepNumber  int
	Title       string
	Description string
	Duration    string
	Links       []roadmapLink
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	careers := []careerSeed{
		softwareEngineer(),
		registeredNurse(),
		electrician(),
		dataScientist(),
		elementaryTeacher(),
		physician(),
		graphicDesigner(),
		accountant(),
		plumber(),
		lawyer(),
		organicFarmer(),
		teaEntrepreneur(),
		coffeeFarmer(),
		poultryFarmer(),
		beekeeper(),
		floriculturist(),
		trekkingGuide(),
		mountaineeringGuide(),
		homestayOperator(),
		hotelManager(),
		tourOperator(),
		chef(),
		travelAgent(),
		bartender(),
		flightAttendant(),
		raftingGuide(),
		itSupport(),
		digitalMarketer(),
		mobileAppDeveloper(),
		freelanceWebDeveloper(),
		ecommerceEntrepreneur(),
		seoSpecialist(),
		cybersecurityAnalyst(),
		droneOperator(),
		pharmacist(),
		labTechnician(),
		ayurvedaDoctor(),
		communityHealthWorker(),
		dentist(),
		radiographer(),
		physiotherapist(),
		yogaInstructor(),
		schoolTeacher(),
		universityProfessor(),
		montessoriTeacher(),
		englishInstructor(),
		onlineTutor(),
		banker(),
		microfinanceOfficer(),
		insuranceAgent(),
		realEstateAgent(),
		cooperativeManager(),
		smallBusinessOwner(),
		stockBroker(),
		remittanceAgent(),
		civilEngineer(),
		electricalEngineer(),
		architect(),
		quantitySurveyor(),
		constructionManager(),
		surveyor(),
		journalist(),
		radioJockey(),
		tvPresenter(),
		photographer(),
		videoEditor(),
		contentWriter(),
		painter(),
		musician(),
		civilServiceOfficer(),
		policeOfficer(),
		judge(),
		diplomat(),
		ngoWorker(),
		socialWorker(),
		carpenter(),
		mason(),
		tailor(),
		evMechanic(),
		motorcycleMechanic(),
		welder(),
		beautyParlor(),
		eventManager(),
		sportsCoach(),
		gymTrainer(),
		supermarketManager(),
		airlinePilot(),
		logisticsManager(),
		publicVehicleDriver(),
		deliveryService(),
		warehouseManager(),
		astrologer(),
		handicraft(),
		cateringService(),
		pashminaBusiness(),
		carpetManufacturer(),
		govtSchoolTeacher(),
		youtuber(),
	}

	log.Printf("Seeding %d base careers...", len(careers))

	// Add bulk-generated careers
	bulkCareers := bulkCareers()
	log.Printf("Adding %d bulk careers...", len(bulkCareers))
	careers = append(careers, bulkCareers...)

	for _, c := range careers {
		// category
		var catID int64
		err := db.Get(&catID, `SELECT id FROM categories WHERE slug = $1`, c.CategorySlug)
		if err != nil {
			err = db.QueryRow(
				`INSERT INTO categories (name, slug, description, icon) VALUES ($1, $2, $3, $4) RETURNING id`,
				c.CategoryName, c.CategorySlug, c.CategoryDesc, c.CategoryIcon,
			).Scan(&catID)
			if err != nil {
				log.Fatalf("insert category %s: %v", c.CategoryName, err)
			}
		}

		tasksJSON, _ := json.Marshal(c.DailyTasks)
		skills := pq.StringArray(c.Skills)

		curr := c.SalaryCurrency
	period := c.SalaryPeriod
	if curr == "" {
		// Old functions had USD salaries - scale down to Nepal NPR reality
		curr = "NPR"
		period = "monthly"
		salMin := c.SalaryMin * 2 / 5
		salMax := c.SalaryMax * 2 / 5
		if salMin < 5000 {
			salMin = 5000
		}
		if salMax < salMin {
			salMax = salMin + 15000
		}
		c.SalaryMin = salMin
		c.SalaryMax = salMax
	} else if period == "" {
		period = "monthly"
	}

	// Build default salary_tiers if not provided
	salaryTiers := c.SalaryTiersJSON
	if salaryTiers == "" {
		entryMin := c.SalaryMin / 2
		entryMax := c.SalaryMin
		avgMin := c.SalaryMin
		avgMax := c.SalaryMax
		expMin := c.SalaryMax
		expMax := c.SalaryMax * 15 / 10
		freelanceMin := avgMin
		freelanceMax := expMax
		salaryTiers = fmt.Sprintf(`{"entry":{"min":%d,"max":%d,"currency":"%s","period":"%s"},"average":{"min":%d,"max":%d,"currency":"%s","period":"%s"},"experienced":{"min":%d,"max":%d,"currency":"%s","period":"%s"},"freelance":{"min":%d,"max":%d,"currency":"%s","period":"%s"}}`,
			entryMin, entryMax, curr, period,
			avgMin, avgMax, curr, period,
			expMin, expMax, curr, period,
			freelanceMin, freelanceMax, curr, period)
	}

	citySalaries := c.CitySalariesJSON
	if citySalaries == "" {
		citySalaries = "{}"
	}

	neContent := c.NepaliContentJSON
	if neContent == "" {
		neContent = "{}"
	}

	demandData := c.DemandDataJSON
	if demandData == "" {
		demandData = `{"trend":"growing","growth_forecast":"stable","opportunities":"moderate","global_opportunity":false,"nepal_demand":"growing"}`
	}

	sourceLabels := c.SourceLabelsJSON
	if sourceLabels == "" {
		sourceLabels = `{"last_updated":"2026-01","salary_source":"Carevo Market Research","demand_source":"Nepal Labor Survey 2025","is_verified":false,"confidence_score":60}`
	}

	metadata := c.MetadataJSON
	if metadata == "" {
		metadata = `{"ai_proof_score":50,"freelance_potential":1,"burnout_risk":3,"remote_potential":1}`
	}

	workLife := c.WorkLifeBalance
	if workLife == 0 {
		workLife = 3
	}
	creativeSc := c.CreativeScore
	if creativeSc == 0 {
		creativeSc = 3
	}
	techSc := c.TechnicalScore
	if techSc == 0 {
		techSc = 3
	}
	freelancePot := c.FreelancePot
	if freelancePot == 0 {
		freelancePot = 1
	}

	var careerID int64
	err = db.QueryRow(
		`INSERT INTO careers (title, slug, summary, description, category_id, daily_tasks, skills,
			salary_min, salary_max, salary_currency, salary_period, difficulty, future_proof_score,
			education_required, outlook, salary_tiers, demand_data, source_labels, career_metadata,
			work_life_balance, study_duration, degree_required, creative_score, technical_score,
			freelance_potential, is_government, is_remote_ok, exam_required, city_salaries, ne_content)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16::jsonb,$17::jsonb,$18::jsonb,$19::jsonb,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29::jsonb,$30::jsonb)
		ON CONFLICT (slug) DO UPDATE SET
			title=EXCLUDED.title, summary=EXCLUDED.summary, description=EXCLUDED.description,
			category_id=EXCLUDED.category_id, daily_tasks=EXCLUDED.daily_tasks, skills=EXCLUDED.skills,
			salary_min=EXCLUDED.salary_min, salary_max=EXCLUDED.salary_max,
			salary_currency=EXCLUDED.salary_currency, salary_period=EXCLUDED.salary_period,
			difficulty=EXCLUDED.difficulty, future_proof_score=EXCLUDED.future_proof_score,
			education_required=EXCLUDED.education_required, outlook=EXCLUDED.outlook,
			salary_tiers=EXCLUDED.salary_tiers, demand_data=EXCLUDED.demand_data,
			source_labels=EXCLUDED.source_labels, career_metadata=EXCLUDED.career_metadata,
			work_life_balance=EXCLUDED.work_life_balance, study_duration=EXCLUDED.study_duration,
			degree_required=EXCLUDED.degree_required, creative_score=EXCLUDED.creative_score,
			technical_score=EXCLUDED.technical_score, freelance_potential=EXCLUDED.freelance_potential,
			is_government=EXCLUDED.is_government, is_remote_ok=EXCLUDED.is_remote_ok,
			exam_required=EXCLUDED.exam_required, city_salaries=EXCLUDED.city_salaries,
			ne_content=EXCLUDED.ne_content
		RETURNING id`,
		c.Title, c.Slug, c.Summary, c.Description, catID, tasksJSON, skills,
		c.SalaryMin, c.SalaryMax, curr, period, c.Difficulty, c.FutureProof,
		c.EducationReq, c.Outlook,
		salaryTiers, demandData, sourceLabels, metadata,
		workLife, c.StudyDuration, c.DegreeRequired, creativeSc, techSc,
		freelancePot, c.IsGovernment, c.IsRemoteOK, c.ExamRequired,
		citySalaries, neContent,
	).Scan(&careerID)
		if err != nil {
			log.Fatalf("insert career %s: %v", c.Title, err)
		}

		for _, tag := range c.Tags {
			db.Exec(`INSERT INTO career_tags (career_id, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING`, careerID, tag)
		}

		for i, r := range c.Resources {
			db.Exec(`INSERT INTO resources (career_id, title, url, description, is_free, sort_order) VALUES ($1,$2,$3,$4,true,$5) ON CONFLICT DO NOTHING`,
				careerID, r.Title, r.URL, r.Description, i)
		}

		for _, s := range c.RoadmapSteps {
			linksJSON, _ := json.Marshal(s.Links)
			db.Exec(`INSERT INTO roadmap_steps (career_id, step_number, title, description, duration, links, sort_order) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (career_id, step_number) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description, duration=EXCLUDED.duration, links=EXCLUDED.links`,
				careerID, s.StepNumber, s.Title, s.Description, s.Duration, linksJSON, s.StepNumber)
		}

		fmt.Printf("seeded: %s\n", c.Title)
	}

	seedSurveyQuestions(db)

	seedAdminUser(db)

	fmt.Println("done!")
}

func seedSurveyQuestions(db *sqlx.DB) {
	questions := []struct {
		SortOrder int
		Category  string
		Text      string
		Options   []string
	}{
		{1, "personality", "Do you enjoy working with people or working alone?", []string{"I love working with people", "I prefer working alone", "I like a mix of both", "Depends on the task"}},
		{2, "personality", "How do you prefer to solve problems?", []string{"Step by step with clear rules", "Creative experimentation", "Research and data analysis", "Ask others for advice"}},
		{3, "personality", "What kind of work environment feels best to you?", []string{"Fast-paced and dynamic", "Quiet and focused", "Outdoors and active", "Collaborative team space"}},
		{4, "personality", "How do you feel about routine tasks?", []string{"I like predictable routines", "I get bored with repetition", "I'm okay with some routine", "I prefer variety every day"}},
		{5, "personality", "Are you comfortable taking risks?", []string{"Yes, I take risks easily", "No, I prefer safety", "Calculated risks only", "Only if there's no other option"}},
		{6, "skills", "Which school subject do you enjoy most?", []string{"Math and numbers", "Science and experiments", "Language and writing", "Arts and crafts"}},
		{7, "skills", "How are your computer skills?", []string{"I can code and build things", "I'm good with software tools", "Basic internet and email", "I prefer hands-on work"}},
		{8, "skills", "Do you enjoy hands-on work with tools or machines?", []string{"Yes, I love fixing and building", "Sometimes, for small tasks", "Not really, I prefer desk work", "Only if needed"}},
		{9, "skills", "How good are you at explaining things to others?", []string{"Very good — I teach naturally", "Good with some practice", "Okay but I get nervous", "I'd rather show than tell"}},
		{10, "skills", "Are you good at organizing and planning?", []string{"Very organized", "Somewhat organized", "I prefer spontaneous", "I struggle with planning"}},
		{11, "interests", "What type of work sounds exciting to you?", []string{"Building or creating something new", "Helping people directly", "Analyzing data and finding patterns", "Leading and managing projects"}},
		{12, "interests", "Would you like to work in an office, outdoors, or remotely?", []string{"Office with a team", "Outdoors and traveling", "Remote from home", "Anywhere is fine"}},
		{13, "interests", "Which industry interests you most?", []string{"Technology and IT", "Healthcare and wellness", "Business and finance", "Creative arts and media"}},
		{14, "interests", "Do you see yourself starting your own business one day?", []string{"Yes, definitely", "Maybe, if the right idea comes", "No, I prefer stable jobs", "I'd rather freelance"}},
		{15, "interests", "What motivates you most in a career?", []string{"High salary and benefits", "Helping society and people", "Creative expression", "Job security and stability"}},
		{16, "work_style", "Can you handle high-stress situations?", []string{"Yes, I stay calm under pressure", "I manage but it affects me", "I prefer low-stress work", "Only in short bursts"}},
		{17, "work_style", "How important is work-life balance to you?", []string{"Very important — I need free time", "Important but I'll work hard", "I don't mind long hours for good pay", "It depends on the stage of life"}},
		{18, "work_style", "Do you like leading teams and making decisions?", []string{"Yes, I'm a natural leader", "Sometimes, if needed", "No, I prefer following", "Only in small groups"}},
		{19, "work_style", "How do you feel about continuing education and learning?", []string{"I love learning new things", "I'll learn if required for my job", "I prefer one-time training", "I'd rather stick to what I know"}},
		{20, "work_style", "Would you prefer a job that lets you travel?", []string{"Yes, travel is important to me", "Occasional travel is fine", "No, I want to stay local", "Remote work is better than travel"}},
	}

	for _, q := range questions {
		optsJSON, _ := json.Marshal(q.Options)
		_, err := db.Exec(
			`INSERT INTO survey_questions (category, question_text, options, sort_order) VALUES ($1, $2, $3, $4)
			 ON CONFLICT DO NOTHING`,
			q.Category, q.Text, optsJSON, q.SortOrder,
		)
		if err != nil {
			log.Printf("seed survey question %d: %v", q.SortOrder, err)
		}
	}
	fmt.Printf("seeded survey questions: %d\n", len(questions))
}

func softwareEngineer() careerSeed {
	return careerSeed{
		CategoryName: "Technology & IT",
		CategorySlug: "technology-it",
		CategoryDesc: "Jobs that involve building, maintaining, and improving software, hardware, and computer systems.",
		CategoryIcon: "💻",
		Title:        "Software Engineer",
		Slug:         "software-engineer",
		Summary:      "Software engineers design, build, and maintain computer programs and apps. They write code, fix bugs, and work with teams to make software that people use every day.",
		Description:  "A software engineer is someone who builds software like apps, websites, and computer programs. They start by understanding what users need, then plan how to build it, write the code, test it to make sure it works, and keep fixing and improving it over time. Software engineers use programming languages like Python, JavaScript, Java, and Go. They work on teams with other engineers, designers, and product managers. Some specialize in frontend (what users see), backend (the hidden server logic), mobile apps, or AI. You can learn without a degree — many engineers are self-taught or went to coding bootcamps. The field changes fast so you need to keep learning new things your whole career.",
		DailyTasks: []string{
			"Write and test code for new features or apps",
			"Fix bugs and errors in existing software",
			"Review other people's code to check for mistakes",
			"Talk with the team about what to build next",
			"Read technical docs to learn new tools",
			"Deploy updates to production servers",
			"Write automated tests to catch problems early",
		},
		Skills: []string{
			"Programming (Python, JavaScript, Java, Go, etc.)",
			"Problem-solving and logical thinking",
			"Understanding of data structures and algorithms",
			"Version control with Git",
			"Database design and SQL",
			"Web development (HTML, CSS, APIs)",
			"Teamwork and communication",
			"Cloud services (AWS, GCP, Azure)",
		},
		SalaryMin:    70000,
		SalaryMax:    160000,
		Difficulty:   4,
		FutureProof:  85,
		EducationReq: "Bachelor's degree in Computer Science or related field, or equivalent self-taught skills / coding bootcamp. Many top engineers are self-taught.",
		Outlook:      "Software engineering is growing much faster than average (25%+ growth). Every company needs software, so jobs are everywhere. AI is changing the field but creating new opportunities rather than replacing engineers.",
		Tags:         []string{"tech", "coding", "remote-work", "high-growth", "engineering"},
		Resources: []resourceSeed{
			{Title: "freeCodeCamp", URL: "https://www.freecodecamp.org", Description: "Free interactive coding lessons - start from zero"},
			{Title: "The Odin Project", URL: "https://www.theodinproject.com", Description: "Free full-stack web development curriculum"},
			{Title: "MIT OpenCourseWare", URL: "https://ocw.mit.edu/search/?q=computer+science", Description: "Free MIT computer science courses"},
			{Title: "Harvard CS50", URL: "https://cs50.harvard.edu", Description: "Free intro to computer science from Harvard"},
			{Title: "LeetCode", URL: "https://leetcode.com", Description: "Practice coding problems for interviews"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Pick a language and build your first tiny thing", Description: "Start with Python or JavaScript — they're the most beginner-friendly. Learn variables (boxes that hold info), loops (repeat stuff), and functions (reusable chunks). Build something tiny but real: a calculator, a to-do list you can add to, or a page that prints 'hello'. Don't try to make it perfect. Just make it work. That feeling when it runs? That's you being a programmer now.", Duration: "2-3 months", Links: []roadmapLink{
				{Title: "freeCodeCamp - Learn to Code for Free", URL: "https://www.freecodecamp.org"},
				{Title: "Harvard CS50 - Intro to Computer Science", URL: "https://cs50.harvard.edu"},
				{Title: "Python for Beginners (Microsoft)", URL: "https://learn.microsoft.com/en-us/training/paths/beginner-python/"},
			}},
			{StepNumber: 2, Title: "Build something real and put it on the internet", Description: "Create a personal website about you. Or a blog. Or a tiny game. Use Git to save versions (like save points in a video game). Push it to GitHub so people can see it. Show a friend and ask what they think. You've leveled up from 'practicing' to 'building.' Feels different, right?", Duration: "2-3 months", Links: []roadmapLink{
				{Title: "GitHub Learning Lab", URL: "https://lab.github.com"},
				{Title: "The Odin Project - Full Stack Curriculum", URL: "https://www.theodinproject.com"},
				{Title: "Free HTML/CSS Course (Codecademy)", URL: "https://www.codecademy.com/learn/learn-html"},
			}},
			{StepNumber: 3, Title: "Teach your app to remember stuff", Description: "Learn SQL and databases. Build a guestbook where people leave messages that stick around. Connect a frontend to a backend to a database — data goes from screen to server to storage and back. That's the magic pipeline behind every app you love. You're building that now.", Duration: "2-3 months", Links: []roadmapLink{
				{Title: "SQL Tutorial (W3Schools)", URL: "https://www.w3schools.com/sql/"},
				{Title: "PostgreSQL Tutorial", URL: "https://www.postgresqltutorial.com"},
				{Title: "REST API Design Tutorial", URL: "https://restfulapi.net"},
			}},
			{StepNumber: 4, Title: "Master one stack by building 2-3 real projects", Description: "Pick one stack (React + Node.js + PostgreSQL is a great choice). Build an e-commerce page with a cart. Build a chat app. Build a dashboard with charts. Each project teaches something the last one didn't. Don't restart — finish. Finished projects are what employers notice.", Duration: "3-4 months", Links: []roadmapLink{
				{Title: "React Tutorial (React.dev)", URL: "https://react.dev/learn"},
				{Title: "Node.js Crash Course (Traversy Media)", URL: "https://www.youtube.com/watch?v=fBNz5xF-Kx4"},
				{Title: "Full Stack Open (University of Helsinki)", URL: "https://fullstackopen.com/en/"},
			}},
			{StepNumber: 5, Title: "Level up with algorithms and system design", Description: "Study data structures (arrays, linked lists, trees, hash tables) and algorithms (sort, search). Practice on LeetCode — start Easy, graduate to Medium. Learn how big systems are designed: URL shorteners, chat apps, Twitter feeds. This is what separates good engineers from great ones.", Duration: "3-4 months", Links: []roadmapLink{
				{Title: "LeetCode - Practice Coding Problems", URL: "https://leetcode.com"},
				{Title: "NeetCode - Algorithm Tutorials", URL: "https://neetcode.io"},
				{Title: "System Design Primer (GitHub)", URL: "https://github.com/donnemartin/system-design-primer"},
			}},
			{StepNumber: 6, Title: "Polish your portfolio and start applying", Description: "Pick your 3 best projects. Clean up the code. Write clear READMEs. Make a simple one-page resume. Apply to 5-10 jobs every week. Do mock interviews. Contribute to open source (even fixing a typo in docs counts!). Every rejection teaches you something. Keep going.", Duration: "2-3 months", Links: []roadmapLink{
				{Title: "Build a Resume (Novoresume)", URL: "https://novoresume.com"},
				{Title: "Mock Interview Practice (Pramp)", URL: "https://www.pramp.com"},
				{Title: "First Timers Only - Open Source", URL: "https://www.firsttimersonly.com"},
			}},
		},
	}
}

func registeredNurse() careerSeed {
	return careerSeed{
		CategoryName: "Healthcare",
		CategorySlug: "healthcare",
		CategoryDesc: "Careers focused on helping people stay healthy, treating illness, and caring for patients.",
		CategoryIcon: "🏥",
		Title:        "Registered Nurse",
		Slug:         "registered-nurse",
		Summary:      "Registered nurses take care of patients in hospitals and clinics. They check vital signs, give medicine, help doctors, and teach people how to stay healthy.",
		Description:  "A registered nurse (RN) is a healthcare worker who takes care of patients. They work in hospitals, doctor's offices, nursing homes, and schools. RNs check patients' blood pressure and temperature, give shots and medicine, help doctors with procedures, and teach families how to care for sick loved ones at home. They spend a lot of time on their feet and need to be kind, calm under pressure, and good at solving problems. Nursing is a very stable career with jobs everywhere because people always get sick and need care.",
		DailyTasks: []string{
			"Check patient vital signs (blood pressure, temperature, heart rate)",
			"Give patients their medicine and treatments",
			"Help doctors during exams and procedures",
			"Clean and bandage wounds",
			"Talk to patients and families about health and recovery",
			"Write down patient information and update charts",
			"Draw blood and collect lab samples",
		},
		Skills: []string{
			"Patient care and bedside manner",
			"Vital signs monitoring",
			"Medication administration",
			"Wound care and first aid",
			"Communication and empathy",
			"Critical thinking under pressure",
			"Teamwork with doctors and other nurses",
			"Medical record keeping",
		},
		SalaryMin:   66000,
		SalaryMax:   135000,
		Difficulty:  4,
		FutureProof: 92,
		EducationReq: "Associate Degree in Nursing (ADN) or Bachelor of Science in Nursing (BSN). Must pass the NCLEX-RN exam to get licensed.",
		Outlook:      "Nursing is growing faster than average (5% growth). There's a shortage of nurses everywhere, so jobs are plentiful. An aging population means even more demand in the future.",
		Tags:         []string{"healthcare", "helping-people", "stable", "hands-on", "shift-work"},
		Resources: []resourceSeed{
			{Title: "Nurses International Free Courses", URL: "https://nursesinternational.org/courses/", Description: "Free nursing courses - fundamentals, pharmacology, pediatrics"},
			{Title: "Nurses Compass", URL: "https://nursescompass.com", Description: "Free NCLEX prep, quizzes, EKG simulator, and nursing study guides"},
			{Title: "RCNi Learning", URL: "https://rcnilearning.com", Description: "Free CPD modules for nurses on many clinical topics"},
			{Title: "NurseJournal Free Resources", URL: "https://nursejournal.org", Description: "Career guides, salary info, and free nursing education articles"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Explore if healthcare is your thing", Description: "Take biology, chemistry, and math seriously in high school. Volunteer at a hospital, nursing home, or clinic. Shadow a nurse if you can. Ask yourself: do I like helping people? Can I stay calm when things get stressful? If yes, you're on the right track. Nursing is tough but incredibly rewarding.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "BLS Registered Nurse Overview", URL: "https://www.bls.gov/ooh/healthcare/registered-nurses.htm"},
				{Title: "Nursing Career Guide (NurseJournal)", URL: "https://nursejournal.org"},
				{Title: "Free Nursing Courses (Nurses International)", URL: "https://nursesinternational.org/courses/"},
			}},
			{StepNumber: 2, Title: "Choose your nursing degree path", Description: "You have two main options: an Associate Degree in Nursing (ADN) takes 2 years at a community college — cheaper and faster. A Bachelor of Science in Nursing (BSN) takes 4 years but opens more doors and usually pays more. Both lead to the same license. Pick the one that fits your life right now.", Duration: "1-3 months", Links: []roadmapLink{
				{Title: "ADN vs BSN Guide", URL: "https://nursejournal.org/degrees/adn-vs-bsn/"},
				{Title: "Find Accredited Nursing Programs", URL: "https://www.acenursing.org/search-programs/"},
				{Title: "Free NCLEX Prep (Nurses Compass)", URL: "https://nursescompass.com"},
			}},
			{StepNumber: 3, Title: "Get through nursing school one semester at a time", Description: "You'll take anatomy, pharmacology, microbiology, and patient care classes. You'll do clinical rotations in real hospitals — that's where you learn the most. It's intense but you're not alone. Form study groups. Use free online resources. Remember why you started.", Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Free Nursing Learning Modules", URL: "https://rcnilearning.com"},
				{Title: "Khan Academy - Anatomy & Physiology", URL: "https://www.khanacademy.org/science/health-and-medicine"},
				{Title: "Nursing Pharmacology Guide", URL: "https://nurseslabs.com/pharmacology/"},
			}},
			{StepNumber: 4, Title: "Pass the NCLEX and become a real RN", Description: "The NCLEX-RN is the big exam at the end. It tests everything you learned. Study 1-2 hours a day for 2-3 months. Use practice questions — lots of them. Take breaks. Trust your training. When you pass (and you will), you're officially a Registered Nurse.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Free NCLEX Practice Questions", URL: "https://nurseslabs.com/nclex-practice-questions/"},
				{Title: "NCLEX Study Guide (Nurses Compass)", URL: "https://nursescompass.com"},
				{Title: "ATI Nursing Test Prep", URL: "https://www.atitesting.com"},
			}},
			{StepNumber: 5, Title: "Land your first nursing job", Description: "Most new nurses start in medical-surgical (med-surg) units — it's the best foundation. Update your resume, practice interview answers, apply everywhere. Hospitals are desperate for nurses. You'll probably get multiple offers. Pick a place with good training and supportive coworkers.", Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Nursing Resume Template", URL: "https://www.indeed.com/career-advice/resume-samples/nurse"},
				{Title: "Indeed Nursing Jobs", URL: "https://www.indeed.com/q-Rn-jobs.html"},
				{Title: "Nurse.org Career Hub", URL: "https://nurse.org"},
			}},
			{StepNumber: 6, Title: "Specialize and grow your career", Description: "After 1-2 years, pick a specialty that excites you — ER (fast-paced), ICU (critical care), pediatrics (kids), surgery (OR), or labor & delivery. Get certified in that specialty. Consider a BSN if you started with an ADN. Go for charge nurse or nurse manager. The ceiling is high.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Nursing Specialty Certification (ANCC)", URL: "https://www.nursingworld.org/our-certifications/"},
				{Title: "RN to BSN Online Programs", URL: "https://nursejournal.org/degrees/rn-to-bsn/"},
				{Title: "Nursing Career Paths Guide", URL: "https://www.allnursingschools.com/registered-nursing/career-paths/"},
			}},
		},
	}
}

func electrician() careerSeed {
	return careerSeed{
		CategoryName: "Skilled Trades",
		CategorySlug: "skilled-trades",
		CategoryDesc: "Hands-on jobs that require technical skills, often learned through apprenticeships. These careers are essential and can't be outsourced.",
		CategoryIcon: "🔧",
		Title:        "Electrician",
		Slug:         "electrician",
		Summary:      "Electricians install and fix electrical wiring in homes, buildings, and factories. They make sure lights, outlets, and machines have safe power to run.",
		Description:  "An electrician is a skilled trade worker who installs, maintains, and repairs electrical systems. They work in homes installing outlets and light fixtures, in offices wiring up computers and lights, and in factories keeping machines running. Electricians read blueprints, follow safety codes, and use tools like wire strippers, multimeters, and conduit benders. The job is hands-on and physical — you climb ladders, crawl in basements, and work in all weather. Most electricians learn through a paid apprenticeship where they earn while they learn. There's always demand because every new building needs wiring and old wiring needs fixing. You don't need college, just a high school diploma and willingness to learn.",
		DailyTasks: []string{
			"Install wiring, outlets, and light fixtures in new buildings",
			"Read blueprints and electrical diagrams",
			"Troubleshoot and fix electrical problems",
			"Test electrical systems with meters and testers",
			"Follow safety rules and building codes",
			"Replace old wiring and upgrade electrical panels",
			"Work with other trades on construction sites",
		},
		Skills: []string{
			"Electrical theory and code knowledge",
			"Blueprint reading",
			"Troubleshooting and problem-solving",
			"Physical stamina and hand-eye coordination",
			"Safety awareness",
			"Customer service (for residential work)",
			"Math for calculating loads and wire sizes",
			"Use of hand and power tools",
		},
		SalaryMin:   40000,
		SalaryMax:   100000,
		Difficulty:  3,
		FutureProof: 88,
		EducationReq: "High school diploma or GED. Then complete a 4-5 year paid apprenticeship program. Must pass a licensing exam to become a journeyman.",
		Outlook:      "Electrician jobs are growing as fast as average (4% growth), but demand is high because many older electricians are retiring. Green energy and EVs are creating new opportunities. The work can't be outsourced or automated.",
		Tags:         []string{"trades", "hands-on", "no-college", "apprenticeship", "outdoor"},
		Resources: []resourceSeed{
			{Title: "SkillCat Free Electrician Training", URL: "https://www.skillcatapp.com", Description: "Free online electrician school with 300+ hours of accredited courses and certifications"},
			{Title: "Alison Free Electrician Courses", URL: "https://alison.com/careers/architecture/electrician", Description: "Free online courses to learn electrical basics and become an electrician"},
			{Title: "OSHA Education Center", URL: "https://www.oshaeducationcenter.com/online-electrician-training-courses", Description: "Online electrician training and safety courses"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Finish high school and see if the trades fit", Description: "Focus on algebra, physics, and any shop or vocational classes. These subjects are the foundation of electrical work. Try a part-time job in construction or maintenance to see if you like working with your hands. Electricians are always in demand — this is a career that can't be shipped overseas or replaced by AI.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "BLS Electrician Career Page", URL: "https://www.bls.gov/ooh/construction-and-extraction/electricians.htm"},
				{Title: "Free Electrician Course (Alison)", URL: "https://alison.com/careers/architecture/electrician"},
				{Title: "Electrical Basics YouTube Playlist", URL: "https://www.youtube.com/results?search_query=electrical+basics+for+beginners"},
			}},
			{StepNumber: 2, Title: "Get into a paid apprenticeship", Description: "An apprenticeship is a 'earn while you learn' program. You work alongside experienced electricians during the day and take classes at night. You get paid from day one. Apply through the International Brotherhood of Electrical Workers (IBEW), local contractors, or trade schools. No college debt needed.", Duration: "1-6 months", Links: []roadmapLink{
				{Title: "Find IBEW Apprenticeships", URL: "https://www.electricaltrainingalliance.org"},
				{Title: "SkillCat Free Electrician Training", URL: "https://www.skillcatapp.com"},
				{Title: "Apprenticeship.gov Search Tool", URL: "https://www.apprenticeship.gov"},
			}},
			{StepNumber: 3, Title: "Learn the trade over 4-5 years", Description: "You'll spend 4-5 years working under a master electrician. You'll learn wiring, circuitry, safety codes, blueprint reading, and how to use tools like multimeters and conduit benders. Take good notes. Ask questions. Every master electrician was once an apprentice too. The hourly pay goes up each year.", Duration: "4-5 years", Links: []roadmapLink{
				{Title: "Free Electrical Code Tutorials", URL: "https://www.oshaeducationcenter.com/online-electrician-training-courses"},
				{Title: "National Electrical Code (NEC) Guide", URL: "https://www.nfpa.org/NEC"},
				{Title: "Electrician's Toolkit App", URL: "https://www.electricianscalculatortool.com"},
			}},
			{StepNumber: 4, Title: "Pass the journeyman exam", Description: "After your apprenticeship, you take your state's journeyman electrician exam. It tests your knowledge of the electrical code, safety, and theory. Study for a few months before taking it. Once you pass, you can work independently without supervision. Your pay jumps significantly.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Journeyman Practice Tests", URL: "https://www.electricianexampractice.com"},
				{Title: "State-by-State License Requirements", URL: "https://www.ncci.com"},
				{Title: "Electrical Code Handbook", URL: "https://www.mikeholt.com"},
			}},
			{StepNumber: 5, Title: "Work as a journeyman and find your specialty", Description: "Now you can work on your own. Pick a path: residential (homes), commercial (offices and stores), or industrial (factories). Each pays differently and has different challenges. Some electricians specialize in solar panels, EV chargers, or home automation. Find what you enjoy most.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Solar Panel Installation Training", URL: "https://www.seia.org"},
				{Title: "EV Charger Installation Guide", URL: "https://www.energy.gov/eere/electricvehicles/ev-charging-installation"},
				{Title: "Home Automation Training", URL: "https://www.coursera.org/courses?query=smart%20home"},
			}},
			{StepNumber: 6, Title: "Go for master electrician or start your own business", Description: "After 2-4 years as a journeyman, take the master electrician exam. Masters can pull permits, run their own business, and charge higher rates. Many master electricians start their own company and make a great living. If you're business-minded, this is the goal.", Duration: "2-4 years", Links: []roadmapLink{
				{Title: "How to Start an Electrical Business", URL: "https://www.ecmag.com"},
				{Title: "Master Electrician Requirements by State", URL: "https://www.ziprecruiter.com/blog/master-electrician-license-requirements/"},
				{Title: "Small Business Administration Guide", URL: "https://www.sba.gov/business-guide"},
			}},
		},
	}
}

func dataScientist() careerSeed {
	return careerSeed{
		CategoryName: "Technology & IT",
		CategorySlug: "technology-it",
		Title:        "Data Scientist",
		Slug:         "data-scientist",
		Summary:      "Data scientists study huge amounts of information to find patterns and help companies make smart decisions. They use math, statistics, and coding.",
		Description:  "A data scientist collects and analyzes large amounts of data to find useful patterns and insights. Companies use these insights to make better decisions — like what products to make, who to market to, or how to improve customer service. Data scientists write code in Python or R, build charts and graphs to show what they find, and use machine learning to predict future trends. They work in almost every industry: tech, healthcare, finance, sports, and more. The field is growing very fast because companies have more data than ever and need help making sense of it.",
		DailyTasks: []string{
			"Clean and organize messy data so it can be analyzed",
			"Build charts and graphs to show data patterns",
			"Write Python or R code to analyze data",
			"Build machine learning models to predict outcomes",
			"Present findings to managers and teams",
			"Run experiments to test ideas",
			"Query databases with SQL to get the right data",
		},
		Skills: []string{
			"Python programming",
			"SQL and database querying",
			"Statistics and probability",
			"Machine learning basics",
			"Data visualization (charts, dashboards)",
			"Critical thinking and curiosity",
			"Communication (explaining data to non-technical people)",
			"Problem-solving with data",
		},
		SalaryMin:   98000,
		SalaryMax:   173000,
		Difficulty:  5,
		FutureProof: 82,
		EducationReq: "Bachelor's degree in Data Science, Computer Science, Statistics, or Mathematics. Many data scientists have a Master's or PhD.",
		Outlook:      "Data science is growing extremely fast (34% growth projected). Almost every industry needs data scientists. AI tools are changing the work but creating more demand for people who understand data deeply.",
		Tags:         []string{"tech", "math", "analytics", "high-growth", "high-salary"},
		Resources: []resourceSeed{
			{Title: "Kaggle Learn", URL: "https://www.kaggle.com/learn", Description: "Free interactive data science and machine learning courses"},
			{Title: "Coursera Data Science (Free Audit)", URL: "https://www.coursera.org/professional-certificates/ibm-data-science", Description: "IBM Data Science Professional Certificate - free to audit"},
			{Title: "StatQuest YouTube", URL: "https://www.youtube.com/@statquest", Description: "Free statistics and machine learning videos explained simply"},
			{Title: "Fast.ai", URL: "https://www.fast.ai", Description: "Free practical deep learning courses for coders"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Start with Python and SQL — the two essentials", Description: "Python is the #1 language for data science. Learn the basics: variables, loops, lists, and dictionaries. Then learn SQL — how to ask questions of a database. Build something tiny: grab a CSV file and write a Python script that reads it and prints interesting facts. You just did data science.", Duration: "2-3 months", Links: []roadmapLink{
				{Title: "Python for Everybody (Free Course)", URL: "https://www.py4e.com"},
				{Title: "SQL Tutorial (SQLZoo)", URL: "https://sqlzoo.net"},
				{Title: "Kaggle Learn - Free Micro Courses", URL: "https://www.kaggle.com/learn"},
			}},
			{StepNumber: 2, Title: "Learn the math that makes data science work", Description: "You don't need to be a math genius. Focus on: descriptive stats (mean, median, standard deviation), probability basics, and hypothesis testing. Use Khan Academy — it explains things like you're 12 (in a good way). Linear algebra helps too but learn it as you go.", Duration: "3-4 months", Links: []roadmapLink{
				{Title: "Khan Academy - Statistics & Probability", URL: "https://www.khanacademy.org/math/statistics-probability"},
				{Title: "StatQuest YouTube Channel", URL: "https://www.youtube.com/@statquest"},
				{Title: "3Blue1Brown - Linear Algebra", URL: "https://www.youtube.com/playlist?list=PLZHQObOWTQDPD3MizzM2xVFitgF8hE_ab"},
			}},
			{StepNumber: 3, Title: "Get comfy with data: clean it, shape it, visualize it", Description: "Learn Pandas (Python library for data) and NumPy (math stuff). Learn to clean messy data — because real data is always messy. Make charts with Matplotlib or Seaborn. Build 2-3 analysis projects: analyze a movie dataset, find trends in weather data, explore what makes houses more expensive.", Duration: "2-3 months", Links: []roadmapLink{
				{Title: "Pandas in 10 Minutes (Official)", URL: "https://pandas.pydata.org/docs/user_guide/10min.html"},
				{Title: "Data Visualization with Python (FreeCodeCamp)", URL: "https://www.freecodecamp.org/learn/data-visualization/"},
				{Title: "Free Datasets to Practice (Kaggle)", URL: "https://www.kaggle.com/datasets"},
			}},
			{StepNumber: 4, Title: "Dive into machine learning", Description: "Learn the two main types: supervised learning (predict a number or category — like predicting house prices or spam) and unsupervised learning (find hidden groups in data). Use scikit-learn — it's the friendliest ML library. Try a Kaggle competition — even placing in the bottom half teaches you tons.", Duration: "3-4 months", Links: []roadmapLink{
				{Title: "Scikit-Learn Tutorials", URL: "https://scikit-learn.org/stable/tutorial/index.html"},
				{Title: "Fast.ai - Practical Deep Learning", URL: "https://www.fast.ai"},
				{Title: "Kaggle Competitions", URL: "https://www.kaggle.com/competitions"},
			}},
			{StepNumber: 5, Title: "Build a portfolio that shows what you can do", Description: "Create 3-5 projects using real datasets. For each one: ask a question, clean the data, analyze it, build a model, make charts, and write a summary of what you found. Put everything on GitHub. Write clear explanations — communication matters as much as the code. Share on LinkedIn.", Duration: "2-3 months", Links: []roadmapLink{
				{Title: "Data Science Project Ideas", URL: "https://www.dataquest.io/blog/data-science-project-ideas/"},
				{Title: "GitHub Portfolio Guide", URL: "https://docs.github.com/en/get-started/quickstart/hello-world"},
				{Title: "Data Science Resume Template", URL: "https://www.novypro.com"},
			}},
			{StepNumber: 6, Title: "Specialize and start your job search", Description: "Pick a focus: NLP (text data), computer vision (images), or MLOps (deploying models). Or pick an industry: healthcare, finance, e-commerce. Apply for data scientist and data analyst roles. Many people start as data analysts and move into data science. The field is growing like crazy (34% growth).", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Data Science Job Board", URL: "https://www.kaggle.com/jobs"},
				{Title: "NLP Course (Hugging Face)", URL: "https://huggingface.co/learn/nlp-course"},
				{Title: "MLOps Basics Guide", URL: "https://ml-ops.org"},
			}},
		},
	}
}

func elementaryTeacher() careerSeed {
	return careerSeed{
		CategoryName: "Education",
		CategorySlug: "education",
		CategoryDesc: "Careers dedicated to teaching and helping students learn and grow.",
		CategoryIcon: "📚",
		Title:        "Elementary School Teacher",
		Slug:         "elementary-school-teacher",
		Summary:      "Elementary teachers help young children learn reading, writing, math, and social skills. They create lesson plans, grade work, and support each child's growth.",
		Description:  "An elementary school teacher works with children in kindergarten through 5th grade. They teach basic subjects like reading, writing, math, science, and social studies. Teachers plan lessons, give assignments, grade work, and track each student's progress. They also manage the classroom, handle behavior issues, and talk with parents about how their children are doing. Teachers spend evenings and weekends planning lessons and grading. It's a rewarding job if you love working with kids and want to make a difference. Summers off and good benefits are big pluses.",
		DailyTasks: []string{
			"Teach reading, writing, and math lessons to the class",
			"Plan fun and educational daily lesson plans",
			"Grade homework, tests, and classwork",
			"Help students who are struggling one-on-one",
			"Talk with parents about student progress",
			"Manage classroom behavior and keep kids safe",
			"Attend staff meetings and training",
		},
		Skills: []string{
			"Patience and empathy with children",
			"Classroom management",
			"Lesson planning and creativity",
			"Communication with parents and staff",
			"Adaptability (every day is different)",
			"Subject knowledge (reading, math, science)",
			"Organization and time management",
			"Conflict resolution",
		},
		SalaryMin:   46000,
		SalaryMax:   90000,
		Difficulty:  3,
		FutureProof: 78,
		EducationReq: "Bachelor's degree in Elementary Education. Must complete a teacher preparation program and pass state licensing exams (like Praxis).",
		Outlook:      "Teaching jobs are growing at an average rate. There are teacher shortages in many areas, especially in rural and low-income schools. Job security is high once you're tenured.",
		Tags:         []string{"education", "children", "helping-people", "summers-off", "stable"},
		Resources: []resourceSeed{
			{Title: "Khan Academy", URL: "https://www.khanacademy.org", Description: "Free teaching resources and tools for educators"},
			{Title: "TeacherTube", URL: "https://www.teachertube.com", Description: "Free educational videos and lesson ideas for teachers"},
			{Title: "ReadWriteThink", URL: "https://www.readwritethink.org", Description: "Free lesson plans and classroom resources for literacy"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "See if you love working with kids", Description: "Volunteer as a tutor, camp counselor, or Sunday school teacher. Babysit or help with local youth programs. Ask yourself: do I have patience? Can I explain things in different ways? Do I genuinely like being around children? Teaching is about connection first, curriculum second.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Teacher Career Overview (BLS)", URL: "https://www.bls.gov/ooh/education-training-and-library/kindergarten-and-elementary-school-teachers.htm"},
				{Title: "A Day in the Life of a Teacher", URL: "https://www.teach.org"},
				{Title: "Volunteer Teaching Opportunities", URL: "https://www.volunteermatch.org"},
			}},
			{StepNumber: 2, Title: "Earn your teaching degree", Description: "Get a bachelor's degree in elementary education. You'll study child development, teaching methods, classroom management, and subject areas like reading and math. You'll also do student teaching — you're placed in a real classroom with a mentor teacher. That's where you learn the most.", Duration: "4 years", Links: []roadmapLink{
				{Title: "Find Accredited Teaching Programs", URL: "https://www.caepnet.org"},
				{Title: "Free Teaching Resources (ReadWriteThink)", URL: "https://www.readwritethink.org"},
				{Title: "Khan Academy for Educators", URL: "https://www.khanacademy.org/khan-for-educators"},
			}},
			{StepNumber: 3, Title: "Get licensed to teach", Description: "Pass your state's teacher certification exams (usually the Praxis series). Apply for your teaching license through the state Department of Education. Requirements vary by state but generally include a degree, student teaching hours, and passing exam scores.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Praxis Exam Prep (ETS)", URL: "https://www.ets.org/praxis/"},
				{Title: "State Teaching License Requirements", URL: "https://www.teach.org/state-requirements"},
				{Title: "Free Praxis Practice Tests", URL: "https://www.240tutoring.com"},
			}},
			{StepNumber: 4, Title: "Find your first teaching job", Description: "Apply to schools in your area. Many new teachers start as substitutes or long-term subs. Look at school district websites, hire fairs, and sites like Indeed. Your first year will be the hardest — you'll work evenings and weekends. It gets easier. Veteran teachers are your best resource.", Duration: "1-6 months", Links: []roadmapLink{
				{Title: "Find Teaching Jobs (SchoolSpring)", URL: "https://www.schoolspring.com"},
				{Title: "New Teacher Survival Guide", URL: "https://www.edutopia.org"},
				{Title: "Teacher Interview Tips", URL: "https://www.indeed.com/career-advice/interviewing/teacher-interview-questions"},
			}},
			{StepNumber: 5, Title: "Consider a master's degree", Description: "Many teachers earn a Master's in Education within their first 5 years. It increases your salary significantly (sometimes $5k-15k more per year) and opens doors to roles like reading specialist, instructional coach, or curriculum developer. Some districts pay for it.", Duration: "2-3 years", Links: []roadmapLink{
				{Title: "Online Master's in Education Programs", URL: "https://www.coursera.org/degrees"},
				{Title: "Teacher Salary Advancement Guide", URL: "https://www.nea.org"},
				{Title: "National Board Certification", URL: "https://www.nbpts.org"},
			}},
			{StepNumber: 6, Title: "Grow into leadership roles", Description: "After a few years, you can become a lead teacher, mentor new teachers, or move into administration (principal, assistant principal). National Board Certification boosts your pay and reputation. Some teachers become curriculum specialists or education consultants. Summers off forever.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Teacher Leadership Roles", URL: "https://www.ascd.org"},
				{Title: "Principal Certification Guide", URL: "https://www.teachercertificationdegrees.com/careers/principal/"},
				{Title: "National Board Certification Info", URL: "https://www.nbpts.org/become-a-candidate/"},
			}},
		},
	}
}

func physician() careerSeed {
	return careerSeed{
		CategoryName: "Healthcare",
		CategorySlug: "healthcare",
		Title:        "Physician (Doctor)",
		Slug:         "physician",
		Summary:      "Doctors diagnose and treat medical conditions. They examine patients, order tests, prescribe medicine, and help people stay healthy.",
		Description:  "Physicians, also called doctors, are medical professionals who diagnose and treat illnesses and injuries. They examine patients, take medical histories, order and interpret tests, prescribe medications, and perform procedures. Doctors work in hospitals, clinics, and private practices. Becoming a doctor takes many years of education — 4 years of college, 4 years of medical school, and 3-7 years of residency training. Doctors earn a high salary but work long hours and deal with a lot of stress. It's a good career if you want to help people, are good at science, and can handle pressure.",
		DailyTasks: []string{
			"Examine patients and listen to their health concerns",
			"Order and interpret lab tests and X-rays",
			"Diagnose illnesses and prescribe treatments",
			"Perform medical procedures",
			"Write prescriptions for medicine",
			"Consult with specialists about patient care",
			"Keep detailed medical records",
			"Talk with patients and families about treatment plans",
		},
		Skills: []string{
			"Scientific knowledge and clinical skills",
			"Diagnostic reasoning and problem-solving",
			"Communication with patients and families",
			"Empathy and bedside manner",
			"Manual dexterity for procedures",
			"Ability to work under pressure",
			"Attention to detail",
			"Teamwork with nurses and other doctors",
		},
		SalaryMin:   200000,
		SalaryMax:   250000,
		Difficulty:  5,
		FutureProof: 90,
		EducationReq: "4-year bachelor's degree, 4 years of medical school (MD or DO), then 3-7 years of residency training. Must pass USMLE exams and get state license.",
		Outlook:      "Doctor jobs are growing as fast as average (3% growth). Demand stays high because of the aging population. Primary care is especially needed in rural areas.",
		Tags:         []string{"healthcare", "helping-people", "high-salary", "long-training", "prestigious"},
		Resources: []resourceSeed{
			{Title: "Khan Academy Medicine", URL: "https://www.khanacademy.org/science/health-and-medicine", Description: "Free medical and health science courses"},
			{Title: "Medscape", URL: "https://www.medscape.com", Description: "Free medical news, clinical information, and CME for doctors"},
			{Title: "NIH MedlinePlus", URL: "https://medlineplus.gov", Description: "Free health information from the US National Library of Medicine"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Build your foundation in high school", Description: "Take biology, chemistry, physics, and math (through calculus). Volunteer at a hospital or clinic. Shadow a doctor if you can. Ask yourself: am I ready for 10+ years of school? Do I handle stress well? If yes, keep going. Medicine is a calling, not just a career.", Duration: "4 years", Links: []roadmapLink{
				{Title: "BLS Physician Overview", URL: "https://www.bls.gov/ooh/healthcare/physicians-and-surgeons.htm"},
				{Title: "Volunteer at a Hospital", URL: "https://www.redcross.org/volunteer"},
				{Title: "Pre-Med High School Guide", URL: "https://students-residents.aamc.org"},
			}},
			{StepNumber: 2, Title: "Get through college with pre-med", Description: "Major in something you enjoy (biology, chemistry, neuroscience) while taking med school prereqs: biology, chemistry, organic chemistry, physics, English. Keep your GPA high (3.5+). Study for and take the MCAT. The MCAT is intense — give yourself 3-6 months of dedicated study.", Duration: "4 years", Links: []roadmapLink{
				{Title: "AAMC Pre-Med Navigator", URL: "https://students-residents.aamc.org/pre-med-navigator"},
				{Title: "Free MCAT Prep (Khan Academy)", URL: "https://www.khanacademy.org/test-prep/mcat"},
				{Title: "Medical School Prerequisites Guide", URL: "https://www.aamc.org"},
			}},
			{StepNumber: 3, Title: "Survive and thrive in medical school", Description: "First 2 years are classroom-heavy: anatomy, pharmacology, pathology, physiology. Last 2 years are clinical rotations where you work in hospitals alongside real doctors. It's intense and exhausting but you're surrounded by people on the same journey. Lean on each other.", Duration: "4 years", Links: []roadmapLink{
				{Title: "Med School Insider Guide", URL: "https://www.usnews.com/education/best-graduate-schools/top-medical-schools"},
				{Title: "Free Med School Resources (Medscape)", URL: "https://www.medscape.com"},
				{Title: "Anki Flashcards for Med School", URL: "https://apps.ankiweb.net"},
			}},
			{StepNumber: 4, Title: "Get through residency — the hardest but most important part", Description: "Residency is 3-7 years of working in a hospital as a doctor-in-training. You'll work long hours (60-80 hour weeks) but you're finally a real doctor. You'll rotate through different specialties. Pick one that fits your personality. The pay is modest but it jumps dramatically after.", Duration: "3-7 years", Links: []roadmapLink{
				{Title: "What to Expect in Residency", URL: "https://www.ama-assn.org"},
				{Title: "Residency Program Directory", URL: "https://www.acgme.org"},
				{Title: "Physician Wellness and Burnout Resources", URL: "https://www.stephanie.xyz"},
			}},
			{StepNumber: 5, Title: "Get board certified in your specialty", Description: "After residency, pass your specialty board exam (internal medicine, pediatrics, surgery, etc.). Get your state medical license. Now you're a fully independent doctor. You can practice, teach, or do research. The average physician salary is $230k+.", Duration: "1 year", Links: []roadmapLink{
				{Title: "Board Certification (ABMS)", URL: "https://www.abms.org"},
				{Title: "State Medical License Requirements", URL: "https://www.fsmb.org"},
				{Title: "Physician Salary by Specialty", URL: "https://www.medscape.com/slideshow/2024-compensation-overview"},
			}},
			{StepNumber: 6, Title: "Build your practice and your life", Description: "Join a hospital, clinic, or start your own practice. Many doctors also teach at medical schools, do research, or write. Your earning potential is high but so is burnout — protect your time, take care of yourself. Consider locum tenens (temporary work) for flexibility.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a Medical Practice", URL: "https://www.ama-assn.org/practice-management"},
				{Title: "Physician Career Options", URL: "https://www.practicelink.com"},
				{Title: "Doctor Wellness and Work-Life Balance", URL: "https://www.stephanie.xyz"},
			}},
		},
	}
}

func graphicDesigner() careerSeed {
	return careerSeed{
		CategoryName: "Arts & Design",
		CategorySlug: "arts-design",
		CategoryDesc: "Creative careers where people use art, color, and design to communicate ideas visually.",
		CategoryIcon: "🎨",
		Title:        "Graphic Designer",
		Slug:         "graphic-designer",
		Summary:      "Graphic designers create visual content like logos, posters, social media graphics, and website layouts. They use color, typography, and images to communicate messages.",
		Description:  "A graphic designer creates visual materials to communicate ideas. They design logos, brochures, social media posts, website layouts, product packaging, and more. They use software like Adobe Photoshop, Illustrator, and Figma. Designers work with clients to understand what they need, then create designs that look good and send the right message. You need an eye for color and composition, creativity, and the ability to take feedback. Many graphic designers are freelancers who work from home. You can learn with free online tools and build a portfolio without a degree.",
		DailyTasks: []string{
			"Create logos, brochures, and marketing materials",
			"Design social media graphics and ads",
			"Edit photos and images in Photoshop",
			"Work with clients to understand their design needs",
			"Make changes based on client feedback",
			"Organize design files and assets",
			"Stay up to date with design trends",
		},
		Skills: []string{
			"Adobe Creative Suite (Photoshop, Illustrator, InDesign)",
			"Color theory and typography",
			"Layout and composition",
			"Creativity and visual thinking",
			"Communication with clients",
			"Time management for deadlines",
			"Basic HTML/CSS (helpful for web design)",
			"Branding and visual identity",
		},
		SalaryMin:   45000,
		SalaryMax:   96000,
		Difficulty:  2,
		FutureProof: 65,
		EducationReq: "Bachelor's degree in Graphic Design or related field is common but not required. A strong portfolio matters more than a degree.",
		Outlook:      "Graphic design jobs grow at an average rate (3%). Competition is strong for the best jobs. AI design tools are changing the field but designers who understand strategy and creativity are still in demand.",
		Tags:         []string{"creative", "design", "freelance", "remote-work", "art"},
		Resources: []resourceSeed{
			{Title: "Coursera Graphic Design (Free Audit)", URL: "https://www.coursera.org/specializations/graphic-design", Description: "Free graphic design specialization from CalArts"},
			{Title: "Canva Design School", URL: "https://www.canva.com/designschool/", Description: "Free design tutorials and courses for beginners"},
			{Title: "Envato Tuts+", URL: "https://tutsplus.com", Description: "Free design tutorials for Photoshop, Illustrator, and more"},
			{Title: "Behance Learning", URL: "https://www.behance.net/learn", Description: "Free design livestreams and Adobe tool tutorials"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn the rules of visual design (then you can break them)", Description: "Study color theory (which colors go together), typography (fonts and how to use them), layout (where things go on a page), and composition (what draws the eye). Canva Design School is free and perfect for beginners. Watch YouTube design channels. The goal isn't to be perfect — it's to build your eye.", Duration: "2-3 months", Links: []roadmapLink{
				{Title: "Canva Design School (Free)", URL: "https://www.canva.com/designschool/"},
				{Title: "CalArts Graphic Design Course (Free Audit)", URL: "https://www.coursera.org/specializations/graphic-design"},
				{Title: "Color Theory for Beginners", URL: "https://www.youtube.com/watch?v=Y6C0J7GHuR4"},
			}},
			{StepNumber: 2, Title: "Get hands-on with design tools", Description: "Learn Adobe Photoshop (photo editing), Illustrator (logos and illustrations), and Figma (UI/web design). If Adobe is too expensive, use free alternatives: GIMP (like Photoshop), Inkscape (like Illustrator), and Canva (easy, browser-based). Pick one and get comfortable first.", Duration: "3-4 months", Links: []roadmapLink{
				{Title: "Free Photoshop Tutorials (Tuts+)", URL: "https://tutsplus.com"},
				{Title: "Figma for Beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=figma+for+beginners"},
				{Title: "GIMP Tutorials (Free Photoshop Alternative)", URL: "https://www.gimp.org/tutorials/"},
			}},
			{StepNumber: 3, Title: "Build a portfolio of fake projects", Description: "Design a logo for a fictional coffee shop. Make a poster for a music festival. Design social media graphics for a fitness brand. Create a menu for a restaurant. Variety shows range. Don't worry about being hired — just make stuff. Your portfolio is your resume in this field.", Duration: "2-3 months", Links: []roadmapLink{
				{Title: "Behance - Design Inspiration & Portfolio", URL: "https://www.behance.net"},
				{Title: "Dribbble - Design Community", URL: "https://dribbble.com"},
				{Title: "Portfolio Examples for Designers", URL: "https://www.flux-academy.com"},
			}},
			{StepNumber: 4, Title: "Learn how to work with clients", Description: "Design is a service business. Practice taking feedback without getting defensive. Learn to write a simple proposal. Research what to charge (start lower, raise as you get better). Create a simple website or Behance profile. Your first few clients will teach you more than any course.", Duration: "1-2 months", Links: []roadmapLink{
				{Title: "How to Price Design Work", URL: "https://shotflow.com"},
				{Title: "Freelance Contract Template (And.co)", URL: "https://www.and.co"},
				{Title: "Client Communication Guide", URL: "https://www.flux-academy.com"},
			}},
			{StepNumber: 5, Title: "Get your first paid design work", Description: "Start on freelancing platforms like Fiverr, Upwork, or 99designs. Offer to design for free for a local nonprofit or friend's business (just ask for a testimonial in return). Your first job might only pay $20. Do great work anyway. Build reviews. Each job leads to the next.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Fiverr - Start Freelancing", URL: "https://www.fiverr.com"},
				{Title: "Upwork for Designers", URL: "https://www.upwork.com"},
				{Title: "Design Freelance Guide (Envato)", URL: "https://design.tutsplus.com"},
			}},
			{StepNumber: 6, Title: "Find your niche and raise your rates", Description: "General designers charge less. Specialists charge more. Pick a niche: logo design, branding, web design, UI/UX, illustration, or packaging. Get really good at that one thing. Raise your rates with every new client. Build long-term relationships. Many designers eventually run their own agency.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "How to Specialize as a Designer", URL: "https://www.smashingmagazine.com"},
				{Title: "UX Design Career Guide", URL: "https://www.springboard.com"},
				{Title: "Running a Design Business", URL: "https://www.designers.how"},
			}},
		},
	}
}

func accountant() careerSeed {
	return careerSeed{
		CategoryName: "Business & Finance",
		CategorySlug: "business-finance",
		CategoryDesc: "Careers that involve managing money, tracking finances, and helping businesses and people make smart financial decisions.",
		CategoryIcon: "💰",
		Title:        "Accountant",
		Slug:         "accountant",
		Summary:      "Accountants keep track of money for companies and people. They prepare tax returns, check financial records, and make sure everything adds up correctly.",
		Description:  "An accountant manages financial records for businesses, organizations, or individuals. They track money coming in and going out, prepare tax returns, create financial reports, and make sure everything follows the law. Accountants work in offices, often for accounting firms, corporations, or government agencies. Some work for themselves. The job requires attention to detail, honesty, and good math skills. You usually need a degree and a CPA (Certified Public Accountant) license to advance. Accounting is a stable career with steady demand because every business needs someone to handle their money.",
		DailyTasks: []string{
			"Record financial transactions in accounting software",
			"Prepare tax returns for individuals or businesses",
			"Check financial records for errors or fraud",
			"Create financial reports like balance sheets",
			"Help clients or managers understand their finances",
			"Reconcile bank statements",
			"Stay updated on tax laws and regulations",
		},
		Skills: []string{
			"Math and analytical skills",
			"Attention to detail and accuracy",
			"Knowledge of accounting software (QuickBooks, Excel)",
			"Understanding of tax laws and regulations",
			"Honesty and ethical judgment",
			"Organization and time management",
			"Communication with clients",
			"Problem-solving",
		},
		SalaryMin:   50000,
		SalaryMax:   120000,
		Difficulty:  3,
		FutureProof: 75,
		EducationReq: "Bachelor's degree in Accounting or Finance. CPA certification is highly recommended for career advancement.",
		Outlook:      "Accounting jobs are growing at an average rate (5%). Every business needs accountants. Automation is changing some tasks but creates demand for higher-level advisory roles.",
		Tags:         []string{"business", "finance", "stable", "office-job", "numbers"},
		Resources: []resourceSeed{
			{Title: "Khan Academy Accounting", URL: "https://www.khanacademy.org/economics-finance-domain/core-finance", Description: "Free accounting and finance courses"},
			{Title: "Coursera Accounting (Free Audit)", URL: "https://www.coursera.org/browse/business/accounting", Description: "Free university-level accounting courses to audit"},
			{Title: "AICPA Free Resources", URL: "https://www.aicpa.org", Description: "Free accounting career resources and guidance from the professional association"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Build your math and business foundation", Description: "Take math, business, and economics classes in high school. Learn Excel — it's the #1 tool for accountants. If you can, take a basic bookkeeping class or watch free YouTube tutorials. Ask yourself: do I enjoy organizing things? Am I detail-oriented? Accounting rewards people who are thorough.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "BLS Accountant Overview", URL: "https://www.bls.gov/ooh/business-and-financial/accountants-and-auditors.htm"},
				{Title: "Excel Tutorials for Beginners (Free)", URL: "https://support.microsoft.com/en-us/office/excel-video-training-9bc05390-e94c-46af-a5b3-d7c22f6990bb"},
				{Title: "Free Bookkeeping Course (AccountingCoach)", URL: "https://www.accountingcoach.com"},
			}},
			{StepNumber: 2, Title: "Get your accounting degree", Description: "Earn a bachelor's in accounting or finance. Take courses in financial accounting, tax, audit, cost accounting, and business law. Do an internship — it's the best way to land your first job. Keep your GPA decent. Join the accounting club. Network with recruiters from Big 4 firms.", Duration: "4 years", Links: []roadmapLink{
				{Title: "Find Accounting Programs (AACSB)", URL: "https://www.aacsb.edu"},
				{Title: "Accounting Internship Guide", URL: "https://www.goingconcern.com"},
				{Title: "Free Accounting Courses (Coursera Audit)", URL: "https://www.coursera.org/browse/business/accounting"},
			}},
			{StepNumber: 3, Title: "Start working and learning on the job", Description: "Most new grads start in public accounting (Deloitte, PwC, EY, KPMG) or as staff accountants for companies. You'll do data entry, reconcile accounts, help with audits, and prepare tax returns. It's not glamorous but it teaches you the fundamentals. Ask questions. Take notes. Be reliable.", Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Big 4 Careers Overview", URL: "https://www.glassdoor.com"},
				{Title: "Resume Tips for Accountants", URL: "https://www.indeed.com/career-advice/resume-samples/accountant"},
				{Title: "Entry-Level Accounting Job Guide", URL: "https://www.roberthalf.com"},
			}},
			{StepNumber: 4, Title: "Get your CPA — it changes everything", Description: "The CPA license is the gold standard. It requires 150 college credits (30 more than a bachelor's), 1-2 years of experience, and passing a 4-part exam. Study for 3-4 months per section. It's hard but worth it — CPAs earn 10-15% more and have way more opportunities.", Duration: "1-2 years", Links: []roadmapLink{
				{Title: "CPA Exam Guide (AICPA)", URL: "https://www.aicpa.org/becomeacpa"},
				{Title: "Free CPA Practice Questions", URL: "https://www.cpareviewforfree.com"},
				{Title: "CPA Requirements by State", URL: "https://nasba.org"},
			}},
			{StepNumber: 5, Title: "Move up: senior, manager, controller", Description: "With 3-5 years of experience and a CPA, you can move to senior accountant, accounting manager, or controller. These roles pay $80k-$150k. You'll manage teams, review work, and advise on big decisions. CPAs can also start their own accounting firm and work for themselves.", Duration: "2-5 years", Links: []roadmapLink{
				{Title: "Accounting Career Paths", URL: "https://www.roberthalf.com"},
				{Title: "How to Start an Accounting Firm", URL: "https://www.journalofaccountancy.com"},
				{Title: "Controller vs CFO Guide", URL: "https://www.cfoselections.com"},
			}},
			{StepNumber: 6, Title: "Specialize and reach executive level", Description: "Focus on a niche: tax, audit, forensic accounting (fraud investigation), or advisory. Some accountants become CFOs. Others get an MBA and move into investment banking or private equity. Accounting is a stable base that can take you in many directions. Every business needs one.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Forensic Accounting Career Guide", URL: "https://www.acfe.com"},
				{Title: "CFO Career Path", URL: "https://www.cfoselections.com"},
				{Title: "MBA for Accountants", URL: "https://www.gmac.com"},
			}},
		},
	}
}

func plumber() careerSeed {
	return careerSeed{
		CategoryName: "Skilled Trades",
		CategorySlug: "skilled-trades",
		Title:        "Plumber",
		Slug:         "plumber",
		Summary:      "Plumbers install and fix pipes that carry water, gas, and waste in homes, offices, and factories. They fix leaks, unclog drains, and install fixtures.",
		Description:  "A plumber installs, maintains, and repairs pipes and fixtures that carry water, gas, and waste. They work in homes fixing leaky faucets and clogged toilets, in new buildings installing pipe systems, and in factories maintaining industrial plumbing. Plumbers read blueprints, use tools like pipe wrenches and soldering torches, and follow building codes. The job is physical — you crawl under houses, work in tight spaces, and sometimes handle messy situations. Like electricians, plumbers learn through paid apprenticeships and don't need a college degree. Plumbing is essential work that can't be replaced by machines or outsourced overseas.",
		DailyTasks: []string{
			"Fix leaky faucets, pipes, and toilets",
			"Unclog drains and sewer lines",
			"Install new water heaters and fixtures",
			"Read blueprints to plan pipe layouts",
			"Cut, bend, and join pipes (copper, PVC, steel)",
			"Test pipe systems for leaks and pressure",
			"Follow building codes and safety rules",
		},
		Skills: []string{
			"Mechanical aptitude and hand tools",
			"Pipe fitting and soldering",
			"Troubleshooting and problem-solving",
			"Physical stamina and strength",
			"Blueprint reading",
			"Customer service and communication",
			"Knowledge of building codes",
			"Math for measuring and calculating",
		},
		SalaryMin:   40000,
		SalaryMax:   97000,
		Difficulty:  3,
		FutureProof: 90,
		EducationReq: "High school diploma or GED. Complete a 4-5 year paid apprenticeship program. Must get state license to work independently.",
		Outlook:      "Plumbing jobs grow at an average rate (4%). Many older plumbers are retiring, creating openings. The work can't be automated or outsourced. New green technologies (water conservation, solar heating) are creating new opportunities.",
		Tags:         []string{"trades", "hands-on", "no-college", "apprenticeship", "essential"},
		Resources: []resourceSeed{
			{Title: "Penn Foster Plumbing Training", URL: "https://www.pennfoster.edu/programs/trades/plumber-career-diploma", Description: "Online plumbing training program (self-paced)"},
			{Title: "SkillCat Plumbing Courses", URL: "https://www.skillcatapp.com", Description: "Free plumbing and HVAC training with certifications"},
			{Title: "BLS Plumber Career Page", URL: "https://www.bls.gov/ooh/construction-and-extraction/plumbers-pipefitters-and-steamfitters.htm", Description: "Free career information, salary data, and job outlook from the government"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Finish high school or get your GED", Description: "You don't need college to be a plumber. Focus on math (especially geometry for measuring pipes), physics, and any shop or vocational classes. If your school offers a trades program, take it. Plumbing is a $60k+ career with zero college debt. That's a huge head start in life.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "BLS Plumber Career Page", URL: "https://www.bls.gov/ooh/construction-and-extraction/plumbers-pipefitters-and-steamfitters.htm"},
				{Title: "Plumbing Career Overview", URL: "https://www.careerone.com.au"},
				{Title: "Trade School vs College Comparison", URL: "https://www.trade-schools.net"},
			}},
			{StepNumber: 2, Title: "Start a paid plumbing apprenticeship", Description: "Apply to a plumbing apprenticeship through a union (UA - United Association), a plumbing company, or a trade school. You'll earn money from day one — usually starting at $15-$25/hour. The apprenticeship lasts 4-5 years. No debt. You get paid to learn. This is the best deal in education.", Duration: "1-6 months", Links: []roadmapLink{
				{Title: "Find Plumbing Apprenticeships", URL: "https://www.apprenticeship.gov"},
				{Title: "United Association (UA) Plumbing", URL: "https://www.ua.org"},
				{Title: "Free Plumbing Courses (SkillCat)", URL: "https://www.skillcatapp.com"},
			}},
			{StepNumber: 3, Title: "Learn the trade on the job", Description: "You'll learn pipe fitting, soldering, drainage, venting, water heater installation, and code requirements. You'll work under experienced plumbers who teach you the tricks of the trade. Take classes at night. Every year your pay goes up. By year 4, you're making good money while still learning.", Duration: "4-5 years", Links: []roadmapLink{
				{Title: "Plumbing 101 Basics (YouTube)", URL: "https://www.youtube.com/results?search_query=plumbing+for+beginners"},
				{Title: "Plumbing Code Essentials", URL: "https://www.iccsafe.org"},
				{Title: "Free Plumbing Training (Penn Foster)", URL: "https://www.pennfoster.edu/programs/trades/plumber-career-diploma"},
			}},
			{StepNumber: 4, Title: "Get your journeyman license", Description: "After completing your apprenticeship, take your state's journeyman plumber exam. It tests your knowledge of plumbing codes, safety, and trade skills. Study for 2-3 months. Once you pass, you can work independently. Your pay jumps to $50k-$80k. Big step.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Journeyman Plumber Exam Prep", URL: "https://www.plumbingexampractice.com"},
				{Title: "State Plumbing License Requirements", URL: "https://www.nicccertified.com"},
				{Title: "Plumber Salary by State", URL: "https://www.fieldedge.com/blog/plumbers-salary/"},
			}},
			{StepNumber: 5, Title: "Work and find your specialty", Description: "You can work in new construction (building new homes/offices), service and repair (fixing leaks and clogs), or commercial/industrial (factories, hospitals). Each pays differently. Specialize in something: trenchless pipe repair, medical gas systems, or green plumbing. Specialists charge more.", Duration: "2-5 years", Links: []roadmapLink{
				{Title: "Plumbing Specialties Guide", URL: "https://www.plumbingcareer.com"},
				{Title: "Green Plumbing Certification", URL: "https://www.greenplumbersusa.com"},
				{Title: "Medical Gas Plumbing Certification", URL: "https://www.medicalgas.org"},
			}},
			{StepNumber: 6, Title: "Go for master plumber or start your business", Description: "After 2+ years as a journeyman, take the master plumber exam. Masters earn $80k-$100k+ and can run their own business. Many plumbers start their own company, hire other plumbers, and build a real asset. Plumbing can't be automated — it's recession-proof and future-proof.", Duration: "2-4 years", Links: []roadmapLink{
				{Title: "How to Start a Plumbing Business", URL: "https://www.fieldedge.com"},
				{Title: "Master Plumber Requirements", URL: "https://www.ziprecruiter.com/blog/master-plumber-license-requirements/"},
				{Title: "SBA Guide to Starting a Business", URL: "https://www.sba.gov/business-guide"},
			}},
		},
	}
}

func lawyer() careerSeed {
	return careerSeed{
		CategoryName: "Legal",
		CategorySlug: "legal",
		CategoryDesc: "Careers in law where professionals help people understand and navigate legal rules, courts, and justice.",
		CategoryIcon: "⚖️",
		Title:        "Lawyer",
		Slug:         "lawyer",
		Summary:      "Lawyers help people and businesses with legal problems. They give advice, write legal documents, and represent clients in court.",
		Description:  "A lawyer, also called an attorney, helps people understand and follow the law. They give legal advice, write contracts and wills, represent clients in court, and help resolve disputes. Lawyers specialize in different areas — some handle criminal cases, others work on divorces, business deals, or injury claims. The job involves a lot of reading, writing, and research. Most lawyers work in law firms, but some work for government, corporations, or run their own practice. Becoming a lawyer requires college, law school (3 years), and passing the bar exam. It's a demanding but respected career with high earning potential.",
		DailyTasks: []string{
			"Research laws and court cases related to client issues",
			"Write legal documents like contracts and court filings",
			"Meet with clients to discuss their legal problems",
			"Negotiate settlements with other lawyers",
			"Represent clients in court hearings",
			"Review contracts and advise on legal risks",
			"Stay current on new laws and regulations",
		},
		Skills: []string{
			"Critical thinking and analysis",
			"Research and writing",
			"Public speaking and argumentation",
			"Negotiation skills",
			"Attention to detail",
			"Understanding of legal procedures",
			"Client counseling and empathy",
			"Ethical judgment and integrity",
		},
		SalaryMin:   72000,
		SalaryMax:   239000,
		Difficulty:  5,
		FutureProof: 72,
		EducationReq: "4-year bachelor's degree, then 3 years of law school (JD degree). Must pass the bar exam in the state where you want to practice.",
		Outlook:      "Lawyer jobs grow at an average rate (5%). Competition for jobs at top firms is intense. Technology is changing how legal work is done but can't replace good legal judgment.",
		Tags:         []string{"legal", "prestigious", "high-salary", "high-stress", "helping-people"},
		Resources: []resourceSeed{
			{Title: "Harvard Free Legal Courses", URL: "https://pll.harvard.edu/subject/law", Description: "Free online law courses from Harvard Law School"},
			{Title: "EdX Law Courses (Free Audit)", URL: "https://www.edx.org/learn/law", Description: "Free university-level law courses to audit"},
			{Title: "Oyez", URL: "https://www.oyez.org", Description: "Free audio and summaries of US Supreme Court cases - great for learning"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Build thinking and speaking skills early", Description: "Take English, history, and government classes. Join debate, mock trial, or student government. These teach you to argue, think on your feet, and write persuasively. Read the news. Pay attention to how laws affect everyday life. Shadow a lawyer if you can. Most lawyers decided in high school.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "BLS Lawyer Overview", URL: "https://www.bls.gov/ooh/legal/lawyers.htm"},
				{Title: "High School Pre-Law Guide", URL: "https://www.pre-law.org"},
				{Title: "Mock Trial Resources", URL: "https://www.nationalmocktrial.org"},
			}},
			{StepNumber: 2, Title: "Pick a college major you enjoy (not just pre-law)", Description: "Law schools accept any major. Popular ones: political science, history, English, philosophy, and criminal justice. Take classes that require lots of reading and writing. Keep your GPA high (3.5+). Build relationships with professors — you'll need recommendation letters for law school.", Duration: "4 years", Links: []roadmapLink{
				{Title: "Pre-Law Major Guide", URL: "https://www.lsac.org/discover-law/pre-law-timeline"},
				{Title: "Free LSAT Prep (Khan Academy)", URL: "https://www.khanacademy.org/prep/lsat"},
				{Title: "Law School Application Timeline", URL: "https://www.lsac.org"},
			}},
			{StepNumber: 3, Title: "Crush the LSAT and apply to law schools", Description: "The LSAT is the most important test you'll take. Study for 3-6 months, take practice tests, and consider a prep course. Your score + GPA determine where you can go. Apply to 8-10 schools: reach, target, and safety. Law school rankings matter but debt matters more. Go where you get scholarships.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Law School Admission Council (LSAC)", URL: "https://www.lsac.org"},
				{Title: "Free LSAT Practice Tests", URL: "https://www.lsac.org/lsat/taking-lsat/lsat-practice-tests"},
				{Title: "Law School Scholarship Guide", URL: "https://www.usnews.com/education/best-graduate-schools/top-law-schools"},
			}},
			{StepNumber: 4, Title: "Get through law school", Description: "First year (1L) of law school is famously brutal. You'll learn case law, legal writing, and the Socratic method (professors cold-call on you). Socratic is nerve-wracking but it teaches you to think fast. In years 2-3, take clinics (real legal work for real clients), internships, and specialized classes. Join law review if you can.", Duration: "3 years", Links: []roadmapLink{
				{Title: "Law School Survival Guide", URL: "https://www.law.uchicago.edu/current-students"},
				{Title: "Free 1L Study Resources (Quimbee)", URL: "https://www.quimbee.com"},
				{Title: "Law School Clinic Guide (ABA)", URL: "https://www.americanbar.org/groups/legal_education/"},
			}},
			{StepNumber: 5, Title: "Pass the bar exam — then celebrate", Description: "The bar exam is a 2-day test covering 14+ areas of law. Most people spend 2-3 months studying full-time (treat it like a full-time job: 8-10 hours/day). Take a bar prep course. Your law school GPA doesn't matter anymore — only whether you pass or fail. 70-80% pass on the first try. You got this.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Bar Exam Guide (NCBE)", URL: "https://www.ncbex.org"},
				{Title: "Bar Prep Courses Comparison", URL: "https://www.top-law-schools.com"},
				{Title: "State Bar Requirements by State", URL: "https://www.ncbex.org/resources/state-bar-requirements/"},
			}},
			{StepNumber: 6, Title: "Start practicing and find your area", Description: "Most new lawyers join a law firm, work for the government (DA, public defender, city attorney), or work in-house for a company. Corporate lawyers make the most money. Public interest lawyers help people. Solo practitioners run their own practice. The beauty of a law degree: you can go in 20+ different directions.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Legal Career Paths (ABA)", URL: "https://www.americanbar.org/careercenter/"},
				{Title: "Solo Practice Starter Guide", URL: "https://www.solosb.com"},
				{Title: "Lawyer Salary by Practice Area", URL: "https://www.nalp.org"},
			}},
		},
	}
}

func organicFarmer() careerSeed {
	return careerSeed{
		CategoryName: "Agriculture & Environment",
		CategorySlug: "agriculture-environment",
		CategoryDesc: "Careers in farming, growing food, and protecting Nepal's natural environment.",
		CategoryIcon: "🌱",
		Title:        "Organic Vegetable Farmer",
		Slug:         "organic-farmer",
		Summary:      "Organic farmers grow vegetables and fruits without using harmful chemicals. They use natural methods to keep soil healthy and grow safe food for their community.",
		Description:  "An organic farmer grows vegetables, fruits, and other crops without using chemical fertilizers or pesticides. In Nepal, organic farming is a growing field because people want healthier food and are willing to pay more for it. Farmers use natural compost, crop rotation, and beneficial insects to protect their crops. They sell at local markets, to hotels and restaurants, or through farmer cooperatives. Many organic farmers in Nepal start small with a few ropanis of land and grow from there. The government of Nepal also supports organic farming through various programs and subsidies.",
		DailyTasks: []string{
			"Prepare soil with compost and natural fertilizers",
			"Plant seeds and transplant seedlings",
			"Water crops and manage irrigation systems",
			"Remove weeds by hand or with natural methods",
			"Harvest vegetables when they are ripe",
			"Pack and transport produce to local markets",
			"Keep records of planting and harvest cycles",
		},
		Skills: []string{
			"Knowledge of organic farming methods",
			"Composting and soil management",
			"Understanding of seasonal planting cycles",
			"Crop rotation and pest management",
			"Basic business and accounting skills",
			"Physical stamina for farm work",
			"Marketing and selling at local markets",
		},
		SalaryMin:   150000,
		SalaryMax:   600000,
		Difficulty:  3,
		FutureProof: 70,
		EducationReq: "No formal degree required. Training available through Agriculture Knowledge Centers and NGOs like Practical Action Nepal. SLC/SEE pass helpful for accessing government programs.",
		Outlook:      "Organic farming in Nepal is growing 10-15% per year. More tourists and middle-class families want organic food. The government offers subsidies for organic certification. Land availability and climate change are challenges but demand keeps rising.",
		Tags:         []string{"agriculture", "organic", "self-employed", "outdoor", "sustainable"},
		Resources: []resourceSeed{
			{Title: "Department of Agriculture Nepal", URL: "https://www.doanepal.gov.np", Description: "Government resources and programs for Nepali farmers"},
			{Title: "Practical Action Nepal Farming", URL: "https://practicalaction.org/where-we-work/nepal/", Description: "Free training and resources for smallholder farmers in Nepal"},
			{Title: "Merojob Agriculture Jobs", URL: "https://www.merojob.com", Description: "Find agriculture and farming jobs in Nepal"},
			{Title: "Kisan Diary Nepal", URL: "https://www.kisandiary.com", Description: "Mobile app and resources for Nepali farmers"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Find a small piece of land to start", Description: "You do not need a big farm to start. Even half a ropani is enough to begin. Look for land near your home or in a village with good water. Talk to other farmers in your area. Ask them what grows well. Start small so you do not lose too much money if something goes wrong.", Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Department of Agriculture Nepal - Land resources", URL: "https://www.doanepal.gov.np"},
				{Title: "Practical Action Nepal - Small farming guide", URL: "https://practicalaction.org/where-we-work/nepal/"},
				{Title: "Kisan Diary - Farming tips", URL: "https://www.kisandiary.com"},
			}},
			{StepNumber: 2, Title: "Learn how to make good compost and healthy soil", Description: "Good soil is the secret to good vegetables. Learn to make compost from cow dung, leaves, and kitchen waste. Do not use chemical fertilizers. They cost money and ruin your soil over time. Visit a local farmer who already farms organically. They will show you for free. Most Nepali farmers love to share knowledge.", Duration: "1-2 months", Links: []roadmapLink{
				{Title: "Department of Agriculture - Composting guide", URL: "https://www.doanepal.gov.np"},
				{Title: "Organic farming training videos Nepal", URL: "https://www.youtube.com/results?search_query=organic+farming+nepal"},
				{Title: "Practical Action - Soil management", URL: "https://practicalaction.org/where-we-work/nepal/"},
			}},
			{StepNumber: 3, Title: "Plant your first crops and care for them every day", Description: "Start with easy vegetables that grow fast: leafy greens, tomatoes, beans, and radishes. Plant in rows and water every morning or evening. Watch for bugs and pick them off by hand instead of using poison. Weeding is boring but very important. Do a little every day so it does not pile up. You will be amazed how much food even a small plot can give you.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Seasonal planting calendar for Nepal", URL: "https://www.doanepal.gov.np"},
				{Title: "Natural pest control methods", URL: "https://www.youtube.com/results?search_query=natural+pest+control+nepal"},
				{Title: "Kisan Diary - Nepal vegetable guide", URL: "https://www.kisandiary.com"},
			}},
			{StepNumber: 4, Title: "Find customers and start selling", Description: "Take your vegetables to the local market or bazaar. Talk to hotels and restaurants in your area — they need fresh vegetables every day. Join a farmer group or cooperative to sell together. Price your vegetables fairly. Tell customers your vegetables are organic — that is a big selling point. Many people in Nepal now ask for organic food.", Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Merojob - Agriculture jobs in Nepal", URL: "https://www.merojob.com"},
				{Title: "Department of Agriculture - Marketing tips", URL: "https://www.doanepal.gov.np"},
				{Title: "Farmer cooperative guide Nepal", URL: "https://www.youtube.com/results?search_query=farming+marketing+nepal"},
			}},
			{StepNumber: 5, Title: "Get organic certification to earn more money", Description: "Organic certification lets you charge higher prices. The Government of Nepal has an organic certification program. Apply through the Department of Agriculture. They will inspect your farm to make sure you are truly organic. Certification costs some money but you will earn it back quickly with better prices.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Nepal organic certification process", URL: "https://www.doanepal.gov.np"},
				{Title: "Benefits of going organic (Nepal)", URL: "https://www.youtube.com/results?search_query=organic+certification+nepal"},
				{Title: "Government subsidies for farmers", URL: "https://www.doanepal.gov.np"},
			}},
			{StepNumber: 6, Title: "Expand your farm and teach others", Description: "Once you are making good money, buy or lease more land. Hire helpers from your village. Grow different vegetables so you have something to sell every season. Build a small greenhouse to grow in winter. Teach other farmers what you learned. Being a farmer in Nepal is not just a job — it is a proud tradition that feeds the nation.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Greenhouse farming in Nepal guide", URL: "https://www.youtube.com/results?search_query=greenhouse+farming+nepal"},
				{Title: "Kisan Diary - Scaling up your farm", URL: "https://www.kisandiary.com"},
				{Title: "Nepal government farming grants", URL: "https://www.doanepal.gov.np"},
			}},
		},
	}
}

func teaEntrepreneur() careerSeed {
	return careerSeed{
		CategoryName: "Agriculture & Environment",
		CategorySlug: "agriculture-environment",
		CategoryIcon: "🍃",
		Title:        "Tea Garden Entrepreneur",
		Slug:         "tea-entrepreneur",
		Summary:      "Tea garden entrepreneurs grow, process, and sell Nepali tea. Nepal's tea is famous worldwide for its quality and unique taste.",
		Description:  "A tea garden entrepreneur runs a business growing and selling tea. Nepal produces some of the finest tea in the world, especially from the eastern regions like Ilam, Dhankuta, and Terhathum. Tea entrepreneurs manage tea gardens, supervise plucking, process the leaves into green, black, or white tea, and sell to local and international markets. Nepal's tea industry is growing fast with more demand from Europe, America, and Japan. You can start with a small garden and expand over time. The Government of Nepal supports tea entrepreneurs through the Nepal Tea and Coffee Development Board.",
		DailyTasks: []string{
			"Supervise tea plucking during harvest seasons",
			"Manage processing of tea leaves (withering, rolling, drying)",
			"Monitor quality of tea at each stage",
			"Keep records of production and sales",
			"Manage workers and pay wages",
			"Build relationships with tea buyers and exporters",
			"Maintain tea plants and garden equipment",
		},
		Skills: []string{
			"Knowledge of tea cultivation and processing",
			"Business management and record keeping",
			"Understanding of tea quality grading",
			"People management and leadership",
			"Marketing and sales skills",
			"Basic financial management",
			"English for international trade",
		},
		SalaryMin:   300000,
		SalaryMax:   2000000,
		Difficulty:  3,
		FutureProof: 75,
		EducationReq: "SLC/SEE pass required. Training from the Nepal Tea and Coffee Development Board. Business courses helpful but not required. Experience working in a tea garden is very valuable.",
		Outlook:      "Nepali tea exports are growing 8-10% every year. More international buyers want single-origin Nepali orthodox tea. The government is investing in tea tourism. Organic and specialty teas (white, oolong) fetch very high prices in international markets.",
		Tags:         []string{"agriculture", "export", "entrepreneur", "nepali-product", "tea"},
		Resources: []resourceSeed{
			{Title: "Nepal Tea and Coffee Development Board", URL: "https://www.teacoffee.gov.np", Description: "Official government body for tea and coffee development in Nepal"},
			{Title: "Himalayan Tea Traders Association", URL: "https://www.nepalteatraders.com", Description: "Resources and networking for Nepali tea entrepreneurs"},
			{Title: "Merojob Agriculture Jobs", URL: "https://www.merojob.com", Description: "Find tea industry jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn about tea from the ground up", Description: "Visit tea gardens in Ilam or Jhapa. Talk to tea farmers and ask them how they grow and process tea. Work in a tea garden for a season to learn the basics. Read about Nepal's tea history. The best tea entrepreneurs started by understanding every step of the process. Free knowledge is everywhere — use it.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Nepal Tea Board - Resources", URL: "https://www.teacoffee.gov.np"},
				{Title: "Tea farming basics (YouTube)", URL: "https://www.youtube.com/results?search_query=tea+farming+nepal"},
				{Title: "Visit Ilam tea gardens", URL: "https://www.welcomenepal.com"},
			}},
			{StepNumber: 2, Title: "Find land and prepare your tea garden", Description: "Tea grows best on hillsides with good drainage at 1000-2000 meters altitude. The eastern hills of Nepal are perfect. Start with 1-2 ropanis of land. Buy good quality tea saplings from the Tea Board or local nurseries. Plant them in rows with proper spacing. Tea plants take 3 years to mature but they will produce for 50+ years. Think of it as planting for your children and grandchildren.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Tea cultivation guide Nepal", URL: "https://www.teacoffee.gov.np"},
				{Title: "Setting up a tea garden (YouTube)", URL: "https://www.youtube.com/results?search_query=tea+garden+nepal+setup"},
				{Title: "Tea plant nursery suppliers", URL: "https://www.teacoffee.gov.np"},
			}},
			{StepNumber: 3, Title: "Buy small processing equipment and start making tea", Description: "You do not need a big factory to start. Small machines for withering, rolling, and drying cost 1-2 lakh rupees. Learn the processing steps: plucking, withering (drying the leaves), rolling (breaking the cells), oxidizing (for black tea), and drying. Each step changes the flavor. Practice until you make consistent quality. Good tea is an art that takes time to learn.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Small scale tea processing equipment", URL: "https://www.youtube.com/results?search_query=small+tea+processing+nepal"},
				{Title: "Tea processing steps explained", URL: "https://www.teacoffee.gov.np"},
				{Title: "Nepal Tea Board training programs", URL: "https://www.teacoffee.gov.np"},
			}},
			{StepNumber: 4, Title: "Find buyers in Nepal and abroad", Description: "Sell your tea locally first — tea shops, hotels, and restaurants in Kathmandu and Pokhara. Then reach out to exporters in Biratnagar and Kathmandu. Attend tea trade fairs. Build a simple website or Facebook page. Nepal's orthodox tea is world-famous. If you make good quality, buyers will find you. Get your tea certified for international standards.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Nepal tea export guide", URL: "https://www.teacoffee.gov.np"},
				{Title: "Nepal Tea Traders Association", URL: "https://www.nepalteatraders.com"},
				{Title: "Facebook marketing for small businesses Nepal", URL: "https://www.youtube.com/results?search_query=facebook+marketing+nepal"},
			}},
			{StepNumber: 5, Title: "Get organic or Fair Trade certification", Description: "International buyers pay much more for certified organic and Fair Trade tea. The certification process takes 1-2 years but is worth it. Your tea can sell for 3-5 times more. The Nepal Tea Board helps with certification costs. Organic tea from Nepal is highly sought after in Europe and North America.", Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Organic certification for Nepali tea", URL: "https://www.teacoffee.gov.np"},
				{Title: "Fair Trade certification process", URL: "https://www.fairtrade.net"},
				{Title: "Exporting Nepali tea guide", URL: "https://www.youtube.com/results?search_query=export+tea+nepal"},
			}},
			{StepNumber: 6, Title: "Expand and build a tea brand", Description: "Once your business is running well, expand your garden and build your own brand. Package your tea with your own label. Create a story about where your tea comes from — people love that. Open a small tea shop in Kathmandu. Consider tea tourism — let visitors see your garden and taste your tea. Nepali tea has a bright future and you can be part of it.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Building a tea brand in Nepal", URL: "https://www.youtube.com/results?search_query=tea+brand+nepal"},
				{Title: "Tea tourism opportunities Nepal", URL: "https://www.welcomenepal.com"},
				{Title: "Nepal Tea Board - Export resources", URL: "https://www.teacoffee.gov.np"},
			}},
		},
	}
}

func coffeeFarmer() careerSeed {
	return careerSeed{
		CategoryName: "Agriculture & Environment",
		CategorySlug: "agriculture-environment",
		CategoryIcon: "☕",
		Title:        "Coffee Farmer / Producer",
		Slug:         "coffee-farmer",
		Summary:      "Coffee farmers grow and process Nepali coffee beans. Nepal's high-altitude coffee is prized for its smooth, mild flavor and is exported worldwide.",
		Description:  "A coffee farmer grows coffee plants, harvests the cherries, and processes them into green coffee beans for roasting. Nepali coffee is special because it grows at high altitudes (800-1600 meters) which gives it a unique smooth taste. Coffee farming started in Nepal in the 1930s and has grown into a major export industry. Farmers in Gulmi, Palpa, Syangja, and Lalitpur districts are leading producers. Coffee plants take 3-4 years to produce their first harvest but then produce for 20+ years. You can grow coffee alongside other crops like ginger and cardamom for extra income.",
		DailyTasks: []string{
			"Care for coffee plants — pruning, mulching, watering",
			"Harvest ripe coffee cherries by hand",
			"Process coffee cherries (pulping, fermenting, washing, drying)",
			"Sort and grade coffee beans by quality",
			"Maintain nursery beds for new coffee plants",
			"Keep records of production and sales",
			"Attend farmer group meetings and training",
		},
		Skills: []string{
			"Coffee cultivation and pruning techniques",
			"Coffee processing methods (washed, natural, honey)",
			"Coffee cupping and quality grading",
			"Basic business management",
			"Understanding of organic farming",
			"Record keeping and basic accounting",
			"Communication and teamwork in farmer groups",
		},
		SalaryMin:   200000,
		SalaryMax:   800000,
		Difficulty:  3,
		FutureProof: 78,
		EducationReq: "No degree required. Training from the Nepal Coffee Development Board or local cooperatives. Experience working on a coffee farm is very helpful. Basic literacy useful for record keeping.",
		Outlook:      "Nepali coffee exports are growing 12-15% annually. International demand for single-origin Himalayan coffee is very high. Climate change is affecting traditional coffee regions but new areas at higher altitudes are opening up. Coffee prices have been steadily rising.",
		Tags:         []string{"agriculture", "export", "coffee", "organic", "nepali-product"},
		Resources: []resourceSeed{
			{Title: "Nepal Coffee Development Board", URL: "https://www.coffee.gov.np", Description: "Government body supporting Nepali coffee farmers"},
			{Title: "National Coffee Growers Association Nepal", URL: "https://www.coffeegrowersnepal.com", Description: "Resources, training, and networking for coffee farmers"},
			{Title: "Kisan Diary Nepal Agriculture", URL: "https://www.kisandiary.com", Description: "Farming tips and resources for Nepali farmers"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn about coffee by visiting a working farm", Description: "Go to Gulmi or Palpa where coffee farming is big. Talk to farmers who grow coffee. Ask them about the challenges and rewards. Work on a coffee farm for a few weeks to understand the work. Read about coffee processing. The Nepal Coffee Board offers free training for new farmers. Most of what you need is free if you are willing to learn.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Nepal Coffee Board training", URL: "https://www.coffee.gov.np"},
				{Title: "Coffee farming basics (YouTube)", URL: "https://www.youtube.com/results?search_query=coffee+farming+nepal"},
				{Title: "Visit coffee farms in Gulmi", URL: "https://www.welcomenepal.com"},
			}},
			{StepNumber: 2, Title: "Get land and plant your first coffee trees", Description: "Coffee likes high altitudes, well-drained soil, and some shade. You can plant coffee on land that is not good for other crops. Start with 500-1000 plants on half a ropani. Get quality saplings from the Coffee Board or local nurseries. Plant during rainy season for best results. Coffee trees need shade, so plant bananas or other trees nearby. Be patient — your first harvest comes in 3 years.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Coffee planting guide Nepal", URL: "https://www.coffee.gov.np"},
				{Title: "Where to buy coffee saplings", URL: "https://www.coffee.gov.np"},
				{Title: "Coffee farm setup tips (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+start+coffee+farm+nepal"},
			}},
			{StepNumber: 3, Title: "Learn to process coffee cherries into beans", Description: "After harvesting, coffee cherries must be processed the same day. The washed process (remove skin, ferment, wash, dry) makes the cleanest tasting coffee. Natural process (dry the whole cherry) makes fruitier coffee. Learn both methods. Buy a small pulping machine (costs 15-30 thousand rupees). Dry beans on raised beds — not on the ground. Good processing makes good coffee.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Coffee processing methods guide", URL: "https://www.coffee.gov.np"},
				{Title: "Small coffee processing equipment Nepal", URL: "https://www.youtube.com/results?search_query=coffee+processing+nepal"},
				{Title: "Coffee quality and cupping guide", URL: "https://www.youtube.com/results?search_query=coffee+cupping+nepal"},
			}},
			{StepNumber: 4, Title: "Join a cooperative and sell your coffee together", Description: "Joining a coffee cooperative helps you get better prices. Cooperatives collect coffee from many small farmers and sell in bulk to exporters. They also provide training, equipment loans, and quality checking. The National Coffee Growers Association can help you find a cooperative near you. Together you get better prices than selling alone.", Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Find coffee cooperatives Nepal", URL: "https://www.coffeegrowersnepal.com"},
				{Title: "Benefits of farmer cooperatives", URL: "https://www.coffee.gov.np"},
				{Title: "Exporting Nepali coffee guide", URL: "https://www.youtube.com/results?search_query=nepali+coffee+export"},
			}},
			{StepNumber: 5, Title: "Improve quality with organic and specialty methods", Description: "Organic coffee sells for 20-30% more. Specialty coffee (scoring 80+ points) sells for even more. Learn cupping (tasting) to check your quality. Keep your farm organic — no chemicals. Plant more shade trees. Pick only ripe cherries (red ones). These small things make a big difference in quality and price. Nepali coffee is already famous — make yours the best.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Organic coffee certification Nepal", URL: "https://www.coffee.gov.np"},
				{Title: "Specialty coffee standards (SCA)", URL: "https://www.sca.coffee"},
				{Title: "Coffee cupping training resources", URL: "https://www.youtube.com/results?search_query=coffee+cupping+tutorial"},
			}},
			{StepNumber: 6, Title: "Expand and build your own coffee brand", Description: "Once you are producing consistently good coffee, create your own brand. Roast small batches and sell directly to cafes in Kathmandu and Pokhara. Package your beans with a story about your farm. Consider opening a small coffee shop that serves your own coffee. Coffee tourism is also growing — let visitors stay on your farm and experience coffee from tree to cup.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a coffee brand in Nepal", URL: "https://www.youtube.com/results?search_query=coffee+brand+nepal"},
				{Title: "Coffee shop business guide", URL: "https://www.merojob.com"},
				{Title: "Coffee tourism in Nepal", URL: "https://www.welcomenepal.com"},
			}},
		},
	}
}

func poultryFarmer() careerSeed {
	return careerSeed{
		CategoryName: "Agriculture & Environment",
		CategorySlug: "agriculture-environment",
		CategoryIcon: "🐔",
		Title:        "Poultry Farmer",
		Slug:         "poultry-farmer",
		Summary:      "Poultry farmers raise chickens for meat or eggs. Chicken meat and eggs are the most widely eaten animal foods in Nepal, so demand is always high.",
		Description:  "A poultry farmer raises chickens for meat (broilers) or eggs (layers). Poultry farming is one of the fastest growing agriculture businesses in Nepal. With the rising population and more people eating chicken and eggs, demand is increasing every year. In Nepal, poultry farms range from small backyard operations with 50 chickens to large commercial farms with 10,000+ birds. You can start small and grow as you learn. The main costs are feed, chicks, medicine, and housing. Poultry farming needs daily attention but can be very profitable if managed well.",
		DailyTasks: []string{
			"Feed and water chickens every morning and evening",
			"Clean chicken houses and remove waste",
			"Check chickens for signs of sickness or injury",
			"Collect eggs several times a day (for layer farms)",
			"Keep records of feed use, mortality, and production",
			"Vaccinate chickens on schedule",
			"Manage temperature and ventilation in the chicken house",
		},
		Skills: []string{
			"Knowledge of chicken breeds and their needs",
			"Basic animal health and vaccination",
			"Farm management and record keeping",
			"Understanding of feed ratios and nutrition",
			"Biosecurity and disease prevention",
			"Business planning and financial management",
			"Physical stamina for daily farm work",
		},
		SalaryMin:   200000,
		SalaryMax:   1200000,
		Difficulty:  3,
		FutureProof: 80,
		EducationReq: "No formal degree needed. Training from the Department of Livestock Services or poultry companies. Experience working on a poultry farm is the best education. Some technical school programs available.",
		Outlook:      "Poultry is one of the fastest growing agriculture sectors in Nepal (10%+ per year). Per capita chicken consumption is rising as incomes grow. The government is promoting poultry farming through subsidies and training programs. Disease outbreaks are a risk but good management reduces this.",
		Tags:         []string{"agriculture", "livestock", "entrepreneur", "food-production", "growing-sector"},
		Resources: []resourceSeed{
			{Title: "Department of Livestock Services Nepal", URL: "https://www.dls.gov.np", Description: "Government resources and programs for livestock farmers"},
			{Title: "Nepal Poultry Guide", URL: "https://www.nepalpoultry.com", Description: "Poultry farming tips, news, and resources for Nepali farmers"},
			{Title: "Merojob Agriculture Jobs", URL: "https://www.merojob.com", Description: "Find poultry farming jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Work on a poultry farm to learn the basics", Description: "Before starting your own farm, work for 3-6 months on someone else's poultry farm. You will learn how to feed chickens, clean houses, spot sick birds, and manage daily operations. This is the best education you can get. You will also learn if poultry farming is really for you. It is hard work with no holidays. But it can be very rewarding.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Find poultry farm jobs in Nepal", URL: "https://www.merojob.com"},
				{Title: "Department of Livestock Services training", URL: "https://www.dls.gov.np"},
				{Title: "Poultry farming basics (YouTube)", URL: "https://www.youtube.com/results?search_query=poultry+farming+nepal+for+beginners"},
			}},
			{StepNumber: 2, Title: "Start small with 100-200 chickens", Description: "Build a simple chicken house with good ventilation and lighting. Buy day-old chicks from a reputable hatchery. Start with broilers (meat chickens) because they grow fast — ready in 6-8 weeks. Feed them commercial chicken feed. Give them clean water and vaccines. Keep the house warm for baby chicks. Your first batch will teach you more than any book.", Duration: "2-3 months", Links: []roadmapLink{
				{Title: "How to build a chicken house", URL: "https://www.youtube.com/results?search_query=poultry+house+design+nepal"},
				{Title: "Hatcheries in Nepal", URL: "https://www.dls.gov.np"},
				{Title: "Chicken feed suppliers Nepal", URL: "https://www.nepalpoultry.com"},
			}},
			{StepNumber: 3, Title: "Learn to prevent and treat common chicken diseases", Description: "Chickens can get sick quickly. Learn to identify common diseases like Ranikhet (Newcastle disease), fowl pox, and coccidiosis. Vaccinate on schedule. Keep the house clean and dry. Separate sick birds immediately. Good biosecurity — not letting outside people or animals in your farm — prevents most problems. A sick chicken can cost you your whole flock. Prevention is cheaper than treatment.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Poultry disease guide for Nepal", URL: "https://www.dls.gov.np"},
				{Title: "Vaccination schedule for chickens", URL: "https://www.nepalpoultry.com"},
				{Title: "Biosecurity for small farms (YouTube)", URL: "https://www.youtube.com/results?search_query=poultry+biosecurity+nepal"},
			}},
			{StepNumber: 4, Title: "Scale up to 500-1000 chickens and make real profit", Description: "Once you have successfully raised 2-3 batches, expand your farm. Build bigger houses. Hire 1-2 helpers. Buy feed in bulk to reduce costs. The profit margin on poultry is small per chicken (50-100 rupees) but volume makes it add up. 1000 broilers at 100 rupees profit each = 100,000 rupees every 2 months. Keep good records of all costs and income.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Scaling up your poultry farm", URL: "https://www.nepalpoultry.com"},
				{Title: "Poultry business financial planning (YouTube)", URL: "https://www.youtube.com/results?search_query=poultry+farm+business+plan+nepal"},
				{Title: "Feed management to reduce costs", URL: "https://www.dls.gov.np"},
			}},
			{StepNumber: 5, Title: "Add layer chickens for steady egg income", Description: "Broilers give you profit in 2 months but layer chickens (egg layers) give you daily income for 1.5-2 years. Add 200-300 layer chickens to your farm. They start laying at 18-20 weeks old. Eggs can be sold to local shops, hotels, and schools. Layer farming is less risky because you get income every day instead of waiting for harvest. Many successful poultry farmers do both broilers and layers.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Layer chicken farming guide Nepal", URL: "https://www.dls.gov.np"},
				{Title: "How to sell eggs in local markets", URL: "https://www.nepalpoultry.com"},
				{Title: "Broiler vs layer comparison (YouTube)", URL: "https://www.youtube.com/results?search_query=broiler+vs+layer+farming+nepal"},
			}},
			{StepNumber: 6, Title: "Become a leader in your local poultry community", Description: "Join the Nepal Poultry Association. Participate in training programs. Teach new farmers what you have learned. Consider adding a feed mill or hatchery to your business. The most successful poultry farmers in Nepal are those who help others succeed too. Poultry farming will always be needed because people love to eat chicken and eggs every day.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Nepal Poultry Association", URL: "https://www.nepalpoultry.com"},
				{Title: "Advanced poultry farming training", URL: "https://www.dls.gov.np"},
				{Title: "Poultry business diversification ideas", URL: "https://www.youtube.com/results?search_query=poultry+farming+business+ideas+nepal"},
			}},
		},
	}
}

func beekeeper() careerSeed {
	return careerSeed{
		CategoryName: "Agriculture & Environment",
		CategorySlug: "agriculture-environment",
		CategoryIcon: "🐝",
		Title:        "Beekeeper / Apiculturist",
		Slug:         "beekeeper",
		Summary:      "Beekeepers raise honeybees to produce honey, beeswax, and other products. Beekeeping also helps farms grow better because bees pollinate crops.",
		Description:  "A beekeeper (also called an apiculturist) manages colonies of honeybees to produce honey, beeswax, royal jelly, and propolis. Beekeeping is an ancient practice in Nepal, especially in the hills where wild honey hunting has been done for centuries. Modern beekeeping uses box hives that are easier to manage. Nepal produces some of the world's most expensive honey — including the famous Himalayan cliff honey from the world's largest honeybee (Apis laboriosa). Beekeeping is also important for pollination — farmers pay beekeepers to bring hives to their fields. You can start with just 5-10 hives and grow from there.",
		DailyTasks: []string{
			"Inspect hives for health, queen activity, and honey stores",
			"Harvest honey when frames are full and capped",
			"Extract honey using a centrifuge or crush method",
			"Feed bees sugar water when flowers are scarce",
			"Manage pests like Varroa mites and wax moths",
			"Prepare hives for winter or migration",
			"Bottle and label honey for sale",
		},
		Skills: []string{
			"Understanding of bee behavior and colony management",
			"Ability to work calmly with bees",
			"Hive construction and maintenance",
			"Honey extraction and processing",
			"Pest and disease identification",
			"Basic business and marketing skills",
			"Knowledge of local flowers and nectar seasons",
		},
		SalaryMin:   150000,
		SalaryMax:   700000,
		Difficulty:  2,
		FutureProof: 82,
		EducationReq: "No degree required. Training from the Nepal Beekeepers Association or Department of Agriculture. Short beekeeping courses offered by NGOs and cooperatives. Practical experience is most important.",
		Outlook:      "Nepali honey exports are growing, especially Himalayan honey which sells for high prices internationally. Beekeeping also supports agriculture through pollination. The government promotes beekeeping through subsidies and training. Demand for organic honey is increasing in Nepal and abroad.",
		Tags:         []string{"agriculture", "honey", "self-employed", "outdoor", "sustainable"},
		Resources: []resourceSeed{
			{Title: "Nepal Beekeepers Association", URL: "https://www.nepalbeekeepers.com", Description: "Resources, training, and networking for Nepali beekeepers"},
			{Title: "Department of Agriculture Nepal", URL: "https://www.doanepal.gov.np", Description: "Government beekeeping programs and subsidies"},
			{Title: "Practical Action Beekeeping Nepal", URL: "https://practicalaction.org/where-we-work/nepal/", Description: "Free beekeeping resources for rural communities"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn from a local beekeeper before buying bees", Description: "Find a beekeeper near you and ask if you can help them for a few weeks. Watch how they open hives, handle bees, and harvest honey. Read free guides online from the Nepal Beekeepers Association. Bees are not hard to keep but you need to understand them. Do not buy bees until you have seen a real hive opened. Many people buy bees and then realize it is not for them.", Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Nepal Beekeepers Association", URL: "https://www.nepalbeekeepers.com"},
				{Title: "Beekeeping for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=beekeeping+for+beginners+nepal"},
				{Title: "Practical Action - Free beekeeping guides", URL: "https://practicalaction.org/where-we-work/nepal/"},
			}},
			{StepNumber: 2, Title: "Start with 5-10 hives and basic equipment", Description: "Buy Langstroth box hives — they are standard and easy to use. Each hive costs about 5,000-8,000 rupees. Buy a smoker, hive tool, and bee suit. Get your first bees from a local beekeeper (they will sell you a nucleus colony or nuc). Place hives near flowering plants but away from people and animals. Morning sun on the hive entrance is good. Set up in spring so bees have time to build up before winter.", Duration: "1-2 months", Links: []roadmapLink{
				{Title: "Where to buy beekeeping equipment in Nepal", URL: "https://www.nepalbeekeepers.com"},
				{Title: "Setting up your first hive (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+set+up+beehive+nepal"},
				{Title: "Choosing a good location for hives", URL: "https://practicalaction.org/where-we-work/nepal/"},
			}},
			{StepNumber: 3, Title: "Learn to inspect hives and manage bees through the seasons", Description: "Open hives every 7-10 days in spring and summer. Look for the queen (she is longer than other bees), eggs (tiny white rice shapes), and honey stores. Add boxes when bees run out of space. In winter, reduce the entrance, make sure they have enough honey, and do not open the hive unless necessary. Each season has different tasks. Keeping bees is about learning the rhythm of the colony.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Seasonal beekeeping calendar Nepal", URL: "https://www.nepalbeekeepers.com"},
				{Title: "How to inspect a hive safely (YouTube)", URL: "https://www.youtube.com/results?search_query=hive+inspection+basics"},
				{Title: "Winter bee management in Nepal", URL: "https://practicalaction.org/where-we-work/nepal/"},
			}},
			{StepNumber: 4, Title: "Harvest your first honey and taste the reward", Description: "Harvest honey when 80% of the frame is capped with wax. Use a bee brush to gently remove bees from frames. Cut the caps with a hot knife. Spin frames in a honey extractor (borrow one at first — they are expensive). Filter the honey through a fine cloth. Let it settle for 2 days. Then bottle it. Your first harvest might be small (5-10 kg per hive) but it will be the sweetest honey you ever tasted. You did that.", Duration: "1-2 months", Links: []roadmapLink{
				{Title: "How to harvest honey correctly (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+harvest+honey"},
				{Title: "Honey extraction and bottling guide", URL: "https://www.nepalbeekeepers.com"},
				{Title: "Where to sell honey in Nepal", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 5, Title: "Grow to 20-50 hives and start selling honey seriously", Description: "With 20 hives you can produce 100-300 kg of honey per year. Sell at local markets, health food stores, and online. Pure organic honey sells for 800-1500 rupees per kg in Nepal. Make a simple label with your name and village. Tell customers your honey is pure and local. Offer tasting at markets. Word of mouth is the best advertising for honey.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Marketing honey in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=selling+honey+nepal"},
				{Title: "Honey quality and purity testing", URL: "https://www.nepalbeekeepers.com"},
				{Title: "Building a honey brand", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 6, Title: "Add pollination services and value-added products", Description: "Farmers will pay you to bring hives to their fields for pollination. This can double your income. Make beeswax candles, lip balms, and propolis tinctures. Try making comb honey or flavored honey (with ginger, lemon, or chili). Teach beekeeping classes to earn extra income. Beekeeping is not just a job — it helps nature and feeds the world.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Pollination services guide", URL: "https://www.nepalbeekeepers.com"},
				{Title: "Value-added bee products (YouTube)", URL: "https://www.youtube.com/results?search_query=beeswax+candles+making+nepal"},
				{Title: "Teaching beekeeping as a business", URL: "https://practicalaction.org/where-we-work/nepal/"},
			}},
		},
	}
}

func floriculturist() careerSeed {
	return careerSeed{
		CategoryName: "Agriculture & Environment",
		CategorySlug: "agriculture-environment",
		CategoryIcon: "🌸",
		Title:        "Floriculture Entrepreneur",
		Slug:         "floriculturist",
		Summary:      "Floriculture entrepreneurs grow and sell flowers, ornamental plants, and decorative arrangements for events, hotels, and homes.",
		Description:  "A floriculture entrepreneur grows flowers and ornamental plants for sale. Nepal's floriculture industry is growing fast because weddings, hotels, and events always need flowers. Marigold, gladiolus, rose, and chrysanthemum are the most popular flowers in Nepal. You can grow flowers in greenhouses or open fields. Many floriculturists also offer decoration services for weddings and parties. The business ranges from small family operations to large commercial nurseries with export potential. Kathmandu Valley, Pokhara, and Chitwan are the main markets.",
		DailyTasks: []string{
			"Plant and care for flower seedlings and plants",
			"Water, fertilize, and prune plants regularly",
			"Harvest flowers at the right stage of bloom",
			"Arrange flowers for events and decorations",
			"Sell flowers at markets or to event planners",
			"Manage greenhouse temperature and humidity",
			"Take orders and coordinate with clients",
		},
		Skills: []string{
			"Knowledge of flower varieties and their growing needs",
			"Greenhouse management skills",
			"Floral arrangement and design",
			"Customer service and client communication",
			"Basic business and financial management",
			"Marketing (especially for weddings and events)",
			"Creativity and aesthetic sense",
		},
		SalaryMin:   200000,
		SalaryMax:   1000000,
		Difficulty:  3,
		FutureProof: 68,
		EducationReq: "SLC/SEE pass recommended. Training available through the Floriculture Association of Nepal and Agriculture Knowledge Centers. Practical experience in a nursery is very valuable. Some agricultural universities offer specialized courses.",
		Outlook:      "Nepal's floriculture industry is growing 10-12% annually. Weddings and festivals (especially Tihar) drive huge demand for flowers. The government provides subsidies for greenhouse construction. Export of cut flowers to India and the Middle East is growing.",
		Tags:         []string{"agriculture", "flowers", "entrepreneur", "creative", "events"},
		Resources: []resourceSeed{
			{Title: "Floriculture Association of Nepal", URL: "https://www.flowernepal.com", Description: "Resources and networking for Nepali floriculture entrepreneurs"},
			{Title: "Department of Agriculture Nepal", URL: "https://www.doanepal.gov.np", Description: "Government programs for floriculture development"},
			{Title: "Merojob Agriculture Jobs", URL: "https://www.merojob.com", Description: "Find floriculture jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Work in a flower nursery to learn the ropes", Description: "Before starting your own business, work in a flower nursery for 3-6 months. You will learn how to propagate plants, care for different flower varieties, and manage a greenhouse. You will also learn what customers want and what sells best. This experience is more valuable than any book or course. Plus you will earn some money while learning.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Find nursery jobs in Nepal", URL: "https://www.merojob.com"},
				{Title: "Floriculture Association of Nepal - Training", URL: "https://www.flowernepal.com"},
				{Title: "Nursery management basics (YouTube)", URL: "https://www.youtube.com/results?search_query=floriculture+nursery+nepal"},
			}},
			{StepNumber: 2, Title: "Start with high-demand flowers like marigold and rose", Description: "Marigold is the most popular flower in Nepal — used for every festival, wedding, and ceremony. Roses are always in demand. Start with these two. Plant in beds or in a small greenhouse. Marigold grows easily from seed and is ready in 2-3 months. Roses take longer but have higher profit margins. Grow what you can sell easily at first.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Marigold farming guide Nepal", URL: "https://www.doanepal.gov.np"},
				{Title: "Rose cultivation in greenhouse (YouTube)", URL: "https://www.youtube.com/results?search_query=rose+farming+nepal"},
				{Title: "Floriculture Association - Crop guides", URL: "https://www.flowernepal.com"},
			}},
			{StepNumber: 3, Title: "Learn floral arrangement and decoration skills", Description: "Flower arrangement is an art. Learn to make bouquets, garlands, centerpieces, and wedding decorations. Watch free tutorials on YouTube. Practice with flowers from your garden. Offer to decorate a friend's wedding or party for free to build your skills and portfolio. Good arrangers charge much more for their work. Decoration services earn more than just selling flowers.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Floral arrangement tutorials (YouTube)", URL: "https://www.youtube.com/results?search_query=floral+arrangement+for+beginners"},
				{Title: "Nepali wedding decoration ideas", URL: "https://www.youtube.com/results?search_query=nepali+wedding+flower+decoration"},
				{Title: "Floriculture Association - Design courses", URL: "https://www.flowernepal.com"},
			}},
			{StepNumber: 4, Title: "Find customers and build relationships", Description: "Approach wedding planners, hotels, and event organizers. Offer free samples of your work. Create a simple Facebook page showing your flowers and arrangements. Go to local markets and sell flowers on Saturdays and during festivals. Tihar is the biggest season for flowers — prepare for it months in advance. Happy customers will call you again.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Facebook marketing for flower businesses Nepal", URL: "https://www.youtube.com/results?search_query=facebook+page+flower+business+nepal"},
				{Title: "Wedding flower packages - Ideas", URL: "https://www.merojob.com"},
				{Title: "Selling at local markets in Nepal", URL: "https://www.youtube.com/results?search_query=selling+flowers+nepal+market"},
			}},
			{StepNumber: 5, Title: "Invest in a greenhouse for year-round production", Description: "A greenhouse lets you grow flowers all year, even in winter. The government subsidizes greenhouse construction through the Prime Minister Agriculture Modernization Project. A small greenhouse (500 sq meters) costs 3-5 lakh rupees with subsidy. With a greenhouse you can grow exotic flowers like orchids and gerbera that earn higher prices. This is a big step but it changes your business completely.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Greenhouse subsidy program Nepal", URL: "https://www.doanepal.gov.np"},
				{Title: "Greenhouse construction guide (YouTube)", URL: "https://www.youtube.com/results?search_query=greenhouse+construction+nepal"},
				{Title: "High-value flower cultivation in greenhouse", URL: "https://www.flowernepal.com"},
			}},
			{StepNumber: 6, Title: "Expand into export and value-added products", Description: "Nepali flowers are in demand in India and the Middle East. Contact export companies. Make value-added products like dried flowers, potpourri, and essential oils from flowers. Open a small flower shop in a good location. Train and hire helpers. The floriculture business in Nepal has huge potential and you can be part of its growth story.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Flower export opportunities Nepal", URL: "https://www.flowernepal.com"},
				{Title: "Value-added flower products (YouTube)", URL: "https://www.youtube.com/results?search_query=dried+flower+craft+ideas+nepal"},
				{Title: "Opening a flower shop in Nepal", URL: "https://www.merojob.com"},
			}},
		},
	}
}

func trekkingGuide() careerSeed {
	return careerSeed{
		CategoryName: "Tourism & Hospitality",
		CategorySlug: "tourism-hospitality",
		CategoryDesc: "Careers in Nepal's tourism industry — guiding visitors, running hotels, and showing the beauty of Nepal to the world.",
		CategoryIcon: "🏔️",
		Title:        "Trekking Guide",
		Slug:         "trekking-guide",
		Summary:      "Trekking guides lead groups on hikes through Nepal's mountains. They keep trekkers safe, share local knowledge, and make sure everyone has a great experience.",
		Description:  "A trekking guide leads groups of trekkers on hiking routes through Nepal's mountains. Guides are responsible for the safety, health, and enjoyment of their group. They plan the route, arrange accommodation in teahouses, manage porters, and share information about the local culture, plants, and wildlife. Famous routes include Everest Base Camp, Annapurna Circuit, Langtang Valley, and the Manaslu Circuit. To be a trekking guide in Nepal, you need a trekking guide license from the Nepal Tourism Board. Guides need to know first aid, speak English (or other languages), and be physically fit. It is a rewarding career that lets you spend every day in the mountains.",
		DailyTasks: []string{
			"Meet trekkers and brief them on the day's hike",
			"Lead the group along trails and check their safety",
			"Call ahead to book teahouses for the night",
			"Manage porters carrying equipment and supplies",
			"Monitor trekkers for altitude sickness symptoms",
			"Share information about mountains, culture, and nature",
			"Handle emergencies like injuries or weather changes",
		},
		Skills: []string{
			"Physical fitness and endurance at high altitude",
			"Knowledge of trekking routes and trails in Nepal",
			"First aid and altitude sickness management",
			"English language (other languages like French, German are a plus)",
			"Navigation using maps, compass, and GPS",
			"Communication and people management",
			"Cultural knowledge of local communities",
		},
		SalaryMin:   300000,
		SalaryMax:   1200000,
		Difficulty:  4,
		FutureProof: 72,
		EducationReq: "SLC/SEE pass minimum. Must complete Nepal Tourism Board trekking guide training and pass license exam. First aid certification required. Higher English proficiency opens better opportunities.",
		Outlook:      "Tourism is a major industry in Nepal and trekking is the top activity for visitors. The government is promoting new trekking routes to spread tourism benefits. Experienced guides can earn good money, especially those speaking multiple languages. Seasonality is a challenge — main seasons are spring (March-May) and autumn (September-November).",
		Tags:         []string{"tourism", "outdoor", "mountains", "fitness", "hospitality"},
		Resources: []resourceSeed{
			{Title: "Nepal Tourism Board - Guide Licensing", URL: "https://www.tourismboard.gov.np", Description: "Official guide licensing and training information"},
			{Title: "Trekking Agencies Association of Nepal (TAAN)", URL: "https://www.taan.org.np", Description: "Resources, training, and networking for trekking professionals"},
			{Title: "Merojob Tourism Jobs", URL: "https://www.merojob.com", Description: "Find trekking guide and tourism jobs in Nepal"},
			{Title: "Mountain Guide Training Nepal", URL: "https://www.youtube.com/results?search_query=trekking+guide+training+nepal", Description: "Free training videos for aspiring trekking guides"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Get fit and learn the trails yourself", Description: "Start by trekking. Do short treks first — Poon Hill, Ghorepani, or Langtang. Then do longer ones like Annapurna Circuit or Everest Base Camp. Trek with a guide first and watch what they do. Learn the names of mountains, villages, and plants along the trail. Take photos and write down directions. The best guides started as trekkers who fell in love with the mountains.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Nepal Tourism Board - Trekking routes", URL: "https://www.tourismboard.gov.np"},
				{Title: "Popular trekking routes in Nepal", URL: "https://www.welcomenepal.com"},
				{Title: "Altitude sickness prevention guide", URL: "https://www.youtube.com/results?search_query=altitude+sickness+prevention"},
			}},
			{StepNumber: 2, Title: "Learn English and basic first aid", Description: "English is essential for trekking guides. Most trekkers speak English. Watch English movies, read books, practice speaking with tourists. Take a basic first aid course from the Red Cross or St. John Ambulance. Learn CPR and how to treat altitude sickness. These skills could save someone's life on the mountain. Many guide training programs include first aid.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Red Cross First Aid Training Nepal", URL: "https://www.nrcs.org"},
				{Title: "English learning resources (free)", URL: "https://www.bbc.co.uk/learningenglish"},
				{Title: "Altitude sickness first aid (YouTube)", URL: "https://www.youtube.com/results?search_query=altitude+sickness+treatment"},
			}},
			{StepNumber: 3, Title: "Complete guide training and get your license", Description: "The Nepal Tourism Board and TAAN (Trekking Agencies Association of Nepal) offer trekking guide training programs. The course covers route knowledge, group management, first aid, environmental awareness, and client service. Pass the written and practical exam to get your official trekking guide license. This license is required by law to work as a guide in Nepal.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "TAAN - Guide training programs", URL: "https://www.taan.org.np"},
				{Title: "Nepal Tourism Board - License requirements", URL: "https://www.tourismboard.gov.np"},
				{Title: "Guide training preparation (YouTube)", URL: "https://www.youtube.com/results?search_query=trekking+guide+license+nepal"},
			}},
			{StepNumber: 4, Title: "Work with a trekking agency to gain experience", Description: "Join a registered trekking agency as a trainee or assistant guide. You will start with simpler routes and smaller groups. Learn from senior guides. Build your reputation for being reliable, knowledgeable, and caring. Agencies prefer guides with good reviews from trekkers. Your first season may be slow but every trek is a learning experience. Be patient and professional.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Find trekking agencies hiring guides", URL: "https://www.merojob.com"},
				{Title: "TAAN - Member agencies list", URL: "https://www.taan.org.np"},
				{Title: "How to be a good trekking guide (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+be+a+good+trekking+guide"},
			}},
			{StepNumber: 5, Title: "Learn a second foreign language", Description: "Guides who speak French, German, Spanish, Chinese, or Japanese earn much more. Many trekking agencies specifically look for foreign language guides. Take language classes in Kathmandu or Pokhara. Practice with tourists. Even basic conversation skills make you more valuable. A guide who speaks English and French can earn 2-3 times more than an English-only guide.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Language classes in Kathmandu", URL: "https://www.merojob.com"},
				{Title: "Free language learning apps (Duolingo)", URL: "https://www.duolingo.com"},
				{Title: "French for trekking guides (YouTube)", URL: "https://www.youtube.com/results?search_query=learn+french+for+tourism"},
			}},
			{StepNumber: 6, Title: "Specialize or start your own agency", Description: "After 3-5 years, choose a specialty: high-altitude mountaineering, cultural tours, bird watching, or photography tours. Many experienced guides start their own small trekking agency. Build relationships with international tour operators. The most successful guides are those who treat every trekker like family. Nepal's mountains will always attract visitors and good guides are always needed.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "How to start a trekking agency in Nepal", URL: "https://www.taan.org.np"},
				{Title: "Niche tourism opportunities in Nepal", URL: "https://www.tourismboard.gov.np"},
				{Title: "International tour operator connections", URL: "https://www.youtube.com/results?search_query=starting+trekking+agency+nepal"},
			}},
		},
	}
}

func mountaineeringGuide() careerSeed {
	return careerSeed{
		CategoryName: "Tourism & Hospitality",
		CategorySlug: "tourism-hospitality",
		CategoryIcon: "⛰️",
		Title:        "Mountaineering Guide / Expedition Leader",
		Slug:         "mountaineering-guide",
		Summary:      "Mountaineering guides lead climbing expeditions on Nepal's highest peaks, including Everest. It is one of the most challenging and respected careers in Nepal.",
		Description:  "A mountaineering guide leads climbing expeditions on Nepal's mountains. They are expert climbers who manage logistics, ensure safety, and help clients reach summits. The most famous destination is Mount Everest (Sagarmatha), but guides also lead climbs on Ama Dablam, Lobuche, Island Peak, and many other peaks. Becoming a mountaineering guide requires years of climbing experience, technical skills in rope work and ice climbing, and certification from the Nepal Mountaineering Association. It is a dangerous but highly respected profession. Guides can earn substantial money during the spring climbing season (April-May).",
		DailyTasks: []string{
			"Check climbing equipment — ropes, carabiners, ice axes, crampons",
			"Fix ropes on climbing routes for clients",
			"Lead clients up the mountain, managing their pace and safety",
			"Monitor weather forecasts and make route decisions",
			"Manage base camp and higher camps (tents, food, oxygen)",
			"Handle medical emergencies at high altitude",
			"Coordinate with expedition agencies and support staff",
		},
		Skills: []string{
			"Advanced rock and ice climbing techniques",
			"High-altitude mountaineering experience (7000m+ peaks)",
			"Rope fixing and route setting on glaciers and steep faces",
			"Knowledge of weather patterns and avalanche assessment",
			"First aid and high-altitude medicine",
			"Expedition logistics and team management",
			"Physical fitness and mental resilience at extreme altitude",
		},
		SalaryMin:   500000,
		SalaryMax:   5000000,
		Difficulty:  5,
		FutureProof: 70,
		EducationReq: "SLC/SEE pass minimum. Must complete Nepal Mountaineering Association training and pass mountaineering guide exams. Advanced climbing courses from international bodies like UIAGM/IFMGA highly valued.",
		Outlook:      "Mountaineering is a cornerstone of Nepal's tourism industry. Everest expeditions bring millions of dollars each year. The number of climbers is increasing. However, it is a dangerous profession with risks of accidents and death. Climate change is making some routes more unstable. Experienced guides with international certifications are in high demand.",
		Tags:         []string{"tourism", "mountains", "extreme", "adventure", "high-risk"},
		Resources: []resourceSeed{
			{Title: "Nepal Mountaineering Association", URL: "https://www.nepalmountaineering.org", Description: "Official body for mountaineering guide certification and training"},
			{Title: "Expedition Operators Association", URL: "https://www.eoanepal.com", Description: "Resources and networking for expedition professionals"},
			{Title: "Mountain Safety Training Nepal", URL: "https://www.youtube.com/results?search_query=mountaineering+training+nepal", Description: "Free mountaineering training videos and safety tips"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Build your climbing foundation on lower peaks", Description: "Start with trekking peaks like Island Peak (6189m), Mera Peak (6476m), and Lobuche East (6119m). These are non-technical or low-technical climbs. Learn basic rope work, crampon walking, and ice axe use. Join a climbing team as a support member or trainee. Build your altitude experience step by step. Never rush altitude — your body needs time to adapt. Every climb teaches you something new.", Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Nepal Mountaineering Association - Peak list", URL: "https://www.nepalmountaineering.org"},
				{Title: "Trekking peak climbing guide (YouTube)", URL: "https://www.youtube.com/results?search_query=island+peak+climbing+guide"},
				{Title: "Altitude training and acclimatization", URL: "https://www.youtube.com/results?search_query=high+altitude+acclimatization+guide"},
			}},
			{StepNumber: 2, Title: "Work on an expedition team to learn the ropes", Description: "Join an Everest or 8000m peak expedition as a kitchen staff, base camp support, or assistant. You will learn how expeditions work: logistics, oxygen systems, camp setup, client management. This experience is invaluable. You will work under senior guides and learn from them. Keep your eyes and ears open. Ask questions when appropriate. Expedition work is hard but you get paid while learning the trade.", Duration: "1-2 seasons", Links: []roadmapLink{
				{Title: "Find expedition jobs in Nepal", URL: "https://www.merojob.com"},
				{Title: "Expedition Operators Association members", URL: "https://www.eoanepal.com"},
				{Title: "Life on an Everest expedition (YouTube)", URL: "https://www.youtube.com/results?search_query=everest+base+camp+expedition+guide"},
			}},
			{StepNumber: 3, Title: "Take advanced mountaineering and rescue courses", Description: "Enroll in the Nepal Mountaineering Association's advanced guide training. Learn crevasse rescue, steep ice climbing, mixed terrain climbing, and avalanche rescue. Get certified in wilderness first responder (WFR) or higher. These certifications are required to work as an official guide. They also keep you and your clients safe. Training is intense but it makes you a professional.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "NMA - Advanced mountaineering guide course", URL: "https://www.nepalmountaineering.org"},
				{Title: "Wilderness First Responder training Nepal", URL: "https://www.nrcs.org"},
				{Title: "Crevasse rescue training (YouTube)", URL: "https://www.youtube.com/results?search_query=crevasse+rescue+techniques"},
			}},
			{StepNumber: 4, Title: "Lead your first expedition on a trekking peak", Description: "After certification, lead climbs on trekking peaks as the head guide. Plan the whole expedition: permits, logistics, staff, equipment, safety plans. Manage your team of assistant guides, cooks, and porters. Take care of your clients' safety and enjoyment. Build your reputation. Good reviews from clients lead to more work. Social media and word of mouth are powerful in the guiding world.", Duration: "1-2 seasons", Links: []roadmapLink{
				{Title: "Peak climbing permit system Nepal", URL: "https://www.nepalmountaineering.org"},
				{Title: "Expedition planning guide (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+plan+an+expedition+nepal"},
				{Title: "Client safety and expedition management", URL: "https://www.eoanepal.com"},
			}},
			{StepNumber: 5, Title: "Climb an 8000m peak to gain elite status", Description: "Climbing an 8000-meter peak like Manaslu, Dhaulagiri, or K2 proves you can operate at the highest level. Many expedition agencies require guides to have summited an 8000m peak before hiring them as head guides for Everest. This step is dangerous and expensive. Train hard, choose your season carefully, and climb with an experienced team. Reaching the summit is a great achievement but coming back safely is the real goal.", Duration: "1-2 seasons", Links: []roadmapLink{
				{Title: "8000m peak climbing guide", URL: "https://www.nepalmountaineering.org"},
				{Title: "High altitude climbing safety tips (YouTube)", URL: "https://www.youtube.com/results?search_query=8000m+peak+climbing+guide"},
				{Title: "Manaslu expedition planning", URL: "https://www.eoanepal.com"},
			}},
			{StepNumber: 6, Title: "Become an IFMGA guide or start your own company", Description: "The highest level of certification is the IFMGA (International Federation of Mountain Guides Associations) pin. IFMGA guides are recognized worldwide and can work anywhere. Many aspirational guides spend years working towards this. Alternatively, start your own expedition company. Build relationships with international clients. The top guides in Nepal earn very well and are respected globally. Remember: in the mountains, there are old guides and bold guides, but very few old bold guides. Stay safe.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "IFMGA guide certification path", URL: "https://www.ifmga.info"},
				{Title: "Starting an expedition company in Nepal", URL: "https://www.eoanepal.com"},
				{Title: "Mountaineering guide career tips (YouTube)", URL: "https://www.youtube.com/results?search_query=mountaineering+guide+career+nepal"},
			}},
		},
	}
}

func homestayOperator() careerSeed {
	return careerSeed{
		CategoryName: "Tourism & Hospitality",
		CategorySlug: "tourism-hospitality",
		CategoryIcon: "🏡",
		Title:        "Homestay Operator",
		Slug:         "homestay-operator",
		Summary:      "Homestay operators welcome tourists into their homes, offering a place to sleep and home-cooked Nepali meals while sharing their culture and daily life.",
		Description:  "A homestay operator runs a small guesthouse in their own home, hosting tourists who want an authentic Nepali experience. Homestays are popular in rural areas and villages near trekking routes. Guests stay in a family's home, eat meals with the family, and experience local life firsthand. The Government of Nepal promotes homestay tourism through the Homestay Registration and Operation Guidelines. Many villages have homestay networks that share guests and set standard prices. Homestay operators earn extra income while sharing their culture. It is especially popular among women, who can earn income without leaving their home.",
		DailyTasks: []string{
			"Prepare clean rooms and beds for arriving guests",
			"Cook breakfast, lunch, and dinner for guests",
			"Greet guests and show them around the village",
			"Organize activities like village walks or cooking lessons",
			"Keep the house clean and comfortable",
			"Manage bookings through phone or online platforms",
			"Handle payments and keep financial records",
		},
		Skills: []string{
			"Cooking traditional Nepali food well",
			"Basic English for communicating with guests",
			"Housekeeping and cleanliness standards",
			"Customer service and hospitality",
			"Basic financial management and record keeping",
			"Knowledge of local culture and attractions",
			"Communication skills and friendliness",
		},
		SalaryMin:   100000,
		SalaryMax:   500000,
		Difficulty:  1,
		FutureProof: 65,
		EducationReq: "No formal education required. Basic literacy helpful for record keeping. Homestay training available through local municipalities and the Nepal Tourism Board. Some basic English is very helpful.",
		Outlook:      "Homestay tourism is growing fast in Nepal. Tourists want authentic experiences, not just hotels. The government is promoting village tourism and homestays in new areas. Homestays near popular trekking routes and national parks do especially well. This is a good option for women and families in rural areas to earn extra income.",
		Tags:         []string{"tourism", "hospitality", "homestay", "rural", "women-friendly"},
		Resources: []resourceSeed{
			{Title: "Nepal Tourism Board - Homestay Guidelines", URL: "https://www.tourismboard.gov.np", Description: "Official homestay registration and operation guidelines"},
			{Title: "Community Homestay Network Nepal", URL: "https://www.communityhomestay.com", Description: "Network connecting homestays with tourists visiting Nepal"},
			{Title: "Facebook Group - Nepali Homestay Operators", URL: "https://www.facebook.com", Description: "Connect with other homestay operators in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Prepare your home for guests", Description: "Clean your home thoroughly. Set aside one or two bedrooms for guests. Make sure the rooms are clean, have good light, and a comfortable bed. Install a clean toilet and bathroom. Simple is fine — tourists want authentic, not luxury. Good hygiene is the most important thing. Ask friends and family what they think. A fresh coat of paint and some local decorations make a big difference.", Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Homestay guidelines Nepal Tourism Board", URL: "https://www.tourismboard.gov.np"},
				{Title: "Setting up a homestay room (YouTube)", URL: "https://www.youtube.com/results?search_query=homestay+setup+ideas+nepal"},
				{Title: "Basic homestay requirements checklist", URL: "https://www.welcomenepal.com"},
			}},
			{StepNumber: 2, Title: "Learn basic English and hospitality skills", Description: "Learn simple English phrases for greeting guests, serving food, and giving directions. Practice with English speakers in your village or nearby town. Learn what tourists expect: clean sheets, hot water, a warm welcome. Be friendly and smiling — Nepali hospitality is world-famous. Take a short hospitality training if available in your area. Many municipalities offer free training.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Basic English for tourism (free lessons)", URL: "https://www.bbc.co.uk/learningenglish"},
				{Title: "Hospitality training videos (YouTube)", URL: "https://www.youtube.com/results?search_query=homestay+hospitality+training+nepal"},
				{Title: "Municipal homestay training programs", URL: "https://www.tourismboard.gov.np"},
			}},
			{StepNumber: 3, Title: "Register your homestay and join a network", Description: "Register your homestay with your local ward office or municipality. Get the official homestay registration certificate from the Nepal Tourism Board. Join the local homestay network or committee in your village. These networks share guests, set prices, and help with marketing. Being part of a network brings more guests than working alone. Together you can also handle bigger groups.", Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Homestay registration process Nepal", URL: "https://www.tourismboard.gov.np"},
				{Title: "Community Homestay Network benefits", URL: "https://www.communityhomestay.com"},
				{Title: "Local homestay committees in Nepal", URL: "https://www.welcomenepal.com"},
			}},
			{StepNumber: 4, Title: "Welcome your first guests and make them happy", Description: "Your first guests are the most important. Go out of your way to make them feel welcome. Greet them at the bus stop. Offer tea and snacks when they arrive. Cook delicious dal bhat. Show them around the village. Ask about their day. A warm welcome leads to good reviews and word of mouth. Many homestays get most of their bookings from recommendations. Make every guest feel like family.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "How to welcome guests at your homestay (YouTube)", URL: "https://www.youtube.com/results?search_query=homestay+host+tips+nepal"},
				{Title: "Nepali cooking for tourists - Recipes", URL: "https://www.youtube.com/results?search_query=cook+dal+bhat+for+tourists"},
				{Title: "Getting good reviews as a homestay", URL: "https://www.communityhomestay.com"},
			}},
			{StepNumber: 5, Title: "List your homestay online to reach more guests", Description: "Get listed on online platforms like Booking.com, Airbnb (if available in Nepal), or specialized homestay websites. Ask a family member or friend with a smartphone to help you take good photos and create a listing. Write a simple description in English. Mention what makes your homestay special — the view, the food, the location. Online listings bring guests from all over the world.", Duration: "1-2 months", Links: []roadmapLink{
				{Title: "List your homestay on Booking.com", URL: "https://www.booking.com"},
				{Title: "Taking good photos for your listing (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+take+photos+for+homestay"},
				{Title: "Community Homestay Network - Listing", URL: "https://www.communityhomestay.com"},
			}},
			{StepNumber: 6, Title: "Expand and offer more experiences", Description: "Add more services to earn extra income. Offer cooking classes where guests learn to make momo, dal bhat, or sel roti. Organize village walks, farm visits, or cultural performances. Plant a vegetable garden to grow food for guests. Build one more guest room. Train family members to help. The best homestays in Nepal offer not just a room but a complete cultural experience. You are not just hosting — you are an ambassador for your village and your country.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Adding experiences to your homestay", URL: "https://www.welcomenepal.com"},
				{Title: "Cultural activities for tourists (YouTube)", URL: "https://www.youtube.com/results?search_query=nepali+cultural+program+homestay"},
				{Title: "Growing your homestay business", URL: "https://www.communityhomestay.com"},
			}},
		},
	}
}

func hotelManager() careerSeed {
	return careerSeed{
		CategoryName: "Tourism & Hospitality",
		CategorySlug: "tourism-hospitality",
		CategoryIcon: "🏨",
		Title:        "Hotel Manager",
		Slug:         "hotel-manager",
		Summary:      "Hotel managers run the daily operations of hotels, lodges, and resorts in Nepal. They make sure guests have a comfortable stay.",
		Description:  "A hotel manager oversees all operations of a hotel, lodge, or resort. They manage staff, handle bookings, ensure cleanliness, manage budgets, and make sure guests are happy. In Nepal, hotel managers work everywhere from small teahouses in the mountains to luxury 5-star hotels in Kathmandu and Pokhara. The job requires skills in management, customer service, finance, and problem-solving. Hotel managers often work long hours, especially during tourist season. Good managers are always in demand because tourism is one of Nepal's biggest industries.",
		DailyTasks: []string{
			"Check in guests and ensure rooms are ready",
			"Supervise housekeeping, kitchen, and front desk staff",
			"Handle guest complaints and solve problems",
			"Manage hotel budgets, expenses, and revenue",
			"Coordinate bookings with travel agencies and online platforms",
			"Inspect rooms, facilities, and food quality",
			"Train new staff and schedule shifts",
		},
		Skills: []string{
			"Hotel operations and management knowledge",
			"Customer service and complaint handling",
			"Staff management and leadership",
			"Financial management and budgeting",
			"English language (other languages a plus)",
			"Hotel booking systems and software",
			"Food and beverage management basics",
		},
		SalaryMin:   400000,
		SalaryMax:   2500000,
		Difficulty:  4,
		FutureProof: 75,
		EducationReq: "Bachelor's degree in Hotel Management or Tourism preferred. Diploma in Hotel Management from institutes like NAAC, KATH, or THM in Nepal. Experience in the hotel industry is very important. Many managers start from entry-level positions.",
		Outlook:      "Nepal's hotel industry is growing with tourism. New hotels and resorts are opening every year. Experienced hotel managers are in high demand, especially those with international experience. The rise of online booking has changed the industry but good management is still essential. Seasonality is a challenge but good managers can find year-round work.",
		Tags:         []string{"tourism", "hospitality", "management", "hotel", "customer-service"},
		Resources: []resourceSeed{
			{Title: "Hotel Association Nepal", URL: "https://www.hotelassociationnepal.org.np", Description: "Resources and networking for hotel professionals in Nepal"},
			{Title: "NAAC Hotel Management College", URL: "https://www.naac.edu.np", Description: "One of Nepal's leading hotel management institutes"},
			{Title: "Merojob Hospitality Jobs", URL: "https://www.merojob.com", Description: "Find hotel and hospitality jobs in Nepal"},
			{Title: "Tourism Human Resource Development", URL: "https://www.tourismboard.gov.np", Description: "Free training programs for hotel and tourism professionals"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Start with any entry-level hotel job", Description: "The best hotel managers started at the bottom. Work as a front desk receptionist, housekeeper, bellboy, or kitchen helper. Learn how each department works. Watch how managers handle guests and staff. Be reliable and hardworking. Promotion comes to those who show initiative. Even a few months of front-line work teaches you more than years of theory. You will understand what your future staff goes through every day.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Hotel jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Hotel Association Nepal - Career resources", URL: "https://www.hotelassociationnepal.org.np"},
				{Title: "Hotel front desk training (YouTube)", URL: "https://www.youtube.com/results?search_query=hotel+front+desk+training"},
			}},
			{StepNumber: 2, Title: "Get formal education in hotel management", Description: "Enroll in a hotel management diploma or degree program. The best institutes in Nepal are NAAC (Nepal Academy of Tourism and Hotel Management), KATH (Kathmandu Academy of Tourism and Hospitality), and THM (Tribhuvan University Institute of Hotel Management). These programs cover front office operations, housekeeping, food and beverage, accounting, and management. A certification from a recognized institute opens doors to better jobs and higher pay.", Duration: "1-4 years", Links: []roadmapLink{
				{Title: "NAAC - Hotel Management courses", URL: "https://www.naac.edu.np"},
				{Title: "KATH - Tourism and hospitality programs", URL: "https://www.kath.edu.np"},
				{Title: "THM - Tribhuvan University hotel management", URL: "https://www.thm.tu.edu.np"},
			}},
			{StepNumber: 3, Title: "Work your way up to supervisor or assistant manager", Description: "After some education, aim for a supervisor role. Front office supervisor, housekeeping supervisor, or food and beverage supervisor. Prove you can manage a team and handle guest issues. Learn the hotel's financial side: how to set room rates, manage occupancy, and control costs. Show your boss you are ready for more responsibility. The jump from supervisor to assistant manager is a big one — prove you are ready for it.", Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Hotel supervisor training (YouTube)", URL: "https://www.youtube.com/results?search_query=hotel+supervisor+training+nepal"},
				{Title: "Hospitality management career path", URL: "https://www.hotelassociationnepal.org.np"},
				{Title: "Hotel revenue management basics", URL: "https://www.youtube.com/results?search_query=hotel+revenue+management+nepal"},
			}},
			{StepNumber: 4, Title: "Become an assistant hotel manager", Description: "As assistant manager, you help run the entire hotel. Manage staff schedules, handle inventory orders, process payroll, and handle guest complaints. You will work closely with the general manager and learn the big picture of hotel operations. This is the training ground for becoming a full manager. Take on extra responsibilities. Volunteer for difficult tasks. Every challenge is a learning opportunity.", Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Hotel Association Nepal - Management training", URL: "https://www.hotelassociationnepal.org.np"},
				{Title: "Hotel management courses (Coursera audit)", URL: "https://www.coursera.org/browse/hospitality"},
				{Title: "Managing hotel staff effectively (YouTube)", URL: "https://www.youtube.com/results?search_query=hotel+management+tips"},
			}},
			{StepNumber: 5, Title: "Get certified in hotel management best practices", Description: "International certifications boost your resume and skills. Consider certifications from AHLA (American Hotel and Lodging Association), or the Hospitality Management certificate from eCornell. Learn about sustainable tourism practices — many hotels in Nepal are going green. Keep up with technology like property management systems (PMS) and online booking platforms. Certified managers earn more and have more job options.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "AHLA certification programs", URL: "https://www.ahla.com"},
				{Title: "eCornell Hospitality Management", URL: "https://www.ecornell.com"},
				{Title: "Sustainable tourism training Nepal", URL: "https://www.tourismboard.gov.np"},
			}},
			{StepNumber: 6, Title: "Become general manager or open your own hotel", Description: "With 5-10 years of experience, you can become a general manager of a hotel or resort. GMs earn good salaries and often get housing and other benefits. Some managers go on to open their own hotel or lodge. This is a big step but very rewarding. Nepal needs more quality accommodation as tourism grows. A good hotel manager is the heart of any successful hotel. Your job is to create a home away from home for every guest.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "How to start a hotel in Nepal", URL: "https://www.hotelassociationnepal.org.np"},
				{Title: "General manager role and responsibilities", URL: "https://www.youtube.com/results?search_query=hotel+general+manager+role"},
				{Title: "Nepal Tourism Board - Hotel investment guide", URL: "https://www.tourismboard.gov.np"},
			}},
		},
	}
}

func tourOperator() careerSeed {
	return careerSeed{
		CategoryName: "Tourism & Hospitality",
		CategorySlug: "tourism-hospitality",
		CategoryIcon: "🚌",
		Title:        "Tour Operator",
		Slug:         "tour-operator",
		Summary:      "Tour operators plan and organize travel packages for tourists visiting Nepal. They arrange transport, accommodation, activities, and guides.",
		Description:  "A tour operator designs and sells travel packages to Nepal. They plan everything: flights or transport, hotels, activities, guides, and meals. They work with travel agencies, airlines, hotels, and guides to create complete travel experiences. Tour operators in Nepal specialize in different types of tours: cultural tours, adventure tours, pilgrimage tours, wildlife safaris, or volunteer travel. You can run your own tour company or work for an existing one. The job requires good planning skills, knowledge of Nepal, and the ability to handle problems quickly. The internet has made it possible for small operators to reach customers worldwide.",
		DailyTasks: []string{
			"Design tour packages based on client requests",
			"Book hotels, transport, and guides for groups",
			"Send itineraries and confirmations to clients",
			"Handle changes and emergencies during tours",
			"Manage budget and track expenses for each tour",
			"Market tours on websites, social media, and to agencies",
			"Follow up with clients after their tour for feedback",
		},
		Skills: []string{
			"In-depth knowledge of Nepal's tourist destinations",
			"Trip planning and itinerary design",
			"Negotiation skills with hotels, transport, and guides",
			"Customer service and communication",
			"English language (other languages a big plus)",
			"Basic accounting and financial management",
			"Marketing and social media skills",
		},
		SalaryMin:   300000,
		SalaryMax:   2000000,
		Difficulty:  3,
		FutureProof: 72,
		EducationReq: "Bachelor's degree in Tourism or Business preferred. Tour operator license from Nepal Tourism Board required. Practical experience in the tourism industry is essential. Many successful operators started as guides or travel agents.",
		Outlook:      "Tourism is one of Nepal's largest industries and is growing. The number of tourists has been increasing year on year (pre-COVID). New destinations and activities are being developed. Online booking makes it easier for small tour operators to find customers. Competition is high but there is room for creative and reliable operators.",
		Tags:         []string{"tourism", "travel", "entrepreneur", "planning", "hospitality"},
		Resources: []resourceSeed{
			{Title: "Nepal Tourism Board - Operator Licensing", URL: "https://www.tourismboard.gov.np", Description: "Official tour operator licensing requirements and information"},
			{Title: "Nepal Association of Tour Operators", URL: "https://www.nato.org.np", Description: "Professional association for tour operators in Nepal"},
			{Title: "Merojob Tourism Jobs", URL: "https://www.merojob.com", Description: "Find tour operator jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Know Nepal inside and out", Description: "Travel across Nepal. Visit the main tourist destinations: Kathmandu Valley, Pokhara, Chitwan, Lumbini, and at least 2-3 trekking regions. Stay in different hotels, eat at different restaurants, use different transport. Take notes. Take photos. The more you know firsthand, the better your tours will be. Tourists ask detailed questions — you need to know the answers. Your personal experience is your greatest asset.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Nepal Tourism Board - Destination guide", URL: "https://www.tourismboard.gov.np"},
				{Title: "Visit Nepal - Official tourism site", URL: "https://www.welcomenepal.com"},
				{Title: "NATO - Nepal tour operator resources", URL: "https://www.nato.org.np"},
			}},
			{StepNumber: 2, Title: "Work for an existing tour operator", Description: "Join an established tour company as a tour coordinator, operations assistant, or reservation agent. Learn how tours are planned, booked, and managed. Understand profit margins, commission structures, and supplier relationships. See how they handle problems when things go wrong. This experience is your practical education. Pay attention to every detail. Good tour operators are masters of details.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Find tour operator jobs in Nepal", URL: "https://www.merojob.com"},
				{Title: "NATO - Member companies list", URL: "https://www.nato.org.np"},
				{Title: "Tour operations training (YouTube)", URL: "https://www.youtube.com/results?search_query=tour+operations+training+nepal"},
			}},
			{StepNumber: 3, Title: "Get your tour operator license", Description: "Register your company with the Company Registrar's Office. Then apply for a tour operator license from the Nepal Tourism Board. The requirements include: minimum capital, office space, and qualified staff. The process takes 2-4 months. Having a proper license makes you legal and builds trust with clients and partners. Many international travel agencies only work with licensed operators.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Nepal Tourism Board - License application", URL: "https://www.tourismboard.gov.np"},
				{Title: "Company registration in Nepal", URL: "https://www.ocr.gov.np"},
				{Title: "Tour operator startup guide (YouTube)", URL: "https://www.youtube.com/results?search_query=start+tour+company+nepal"},
			}},
			{StepNumber: 4, Title: "Build relationships with suppliers and partners", Description: "Good relationships with hotels, guides, transport companies, and airlines are the key to a successful tour business. Negotiate rates. Build a network of reliable partners you can trust. Visit properties and meet people in person. A handshake and a personal relationship matter a lot in Nepal. Keep a list of trusted partners for every budget level — budget, mid-range, and luxury.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Hotel Association Nepal - Partner directory", URL: "https://www.hotelassociationnepal.org.np"},
				{Title: "Nepal Association of Tour Operators - Networking", URL: "https://www.nato.org.np"},
				{Title: "Building tourism partnerships (YouTube)", URL: "https://www.youtube.com/results?search_query=tourism+business+networking+nepal"},
			}},
			{StepNumber: 5, Title: "Create a website and start marketing", Description: "Build a simple, professional website showing your tour packages. Use good photos (hire a photographer if needed). Write clear descriptions. Accept online payments through bank transfer or payment gateways. Create social media pages — Instagram and Facebook are powerful for tourism. Share photos and videos of your tours. Encourage happy clients to leave reviews on TripAdvisor and Google. Online presence brings customers from around the world.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Website builder for small business", URL: "https://www.wix.com"},
				{Title: "Social media marketing for tourism (YouTube)", URL: "https://www.youtube.com/results?search_query=tourism+marketing+social+media+nepal"},
				{Title: "Getting reviews on TripAdvisor", URL: "https://www.tripadvisor.com"},
			}},
			{StepNumber: 6, Title: "Specialize and grow your business", Description: "Find your niche: cultural tours, photography tours, food tours, bird watching, yoga retreats, or volunteer travel. Specialized tours command higher prices. Hire more staff as you grow. Open a small office. Attend international travel trade shows like ITB Berlin or SATTE India. The most successful tour operators in Nepal are those who constantly learn, adapt, and take care of their clients like family. Tourism is not just business in Nepal — it is hospitality from the heart.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Niche tourism opportunities Nepal", URL: "https://www.tourismboard.gov.np"},
				{Title: "International travel trade shows", URL: "https://www.itb.com"},
				{Title: "Scaling your tour business (YouTube)", URL: "https://www.youtube.com/results?search_query=grow+tour+company+nepal"},
			}},
		},
	}
}

func chef() careerSeed {
	return careerSeed{
		CategoryName: "Tourism & Hospitality",
		CategorySlug: "tourism-hospitality",
		CategoryIcon: "👨‍🍳",
		Title:        "Chef / Cook",
		Slug:         "chef",
		Summary:      "Chefs prepare and cook food in hotels, restaurants, and cafes across Nepal. They create menus, manage kitchens, and make sure food tastes delicious.",
		Description:  "A chef is a trained professional cook who manages the kitchen and prepares food. In Nepal, chefs work in hotels, restaurants, cafes, trekking lodges, and even expedition base camps. Nepali cuisine is famous worldwide — dal bhat, momo, sel roti, and many more dishes. Chefs can specialize in Nepali cuisine, Indian, Chinese, Continental, or fusion cooking. The career path goes from kitchen helper (commis) to line cook, to sous chef, to head chef (executive chef). Chefs need to be creative, hardworking, and able to handle pressure during busy hours. Nepal's growing restaurant scene offers many opportunities.",
		DailyTasks: []string{
			"Prepare ingredients — chopping vegetables, marinating meat",
			"Cook dishes according to recipes and standards",
			"Plate and present dishes attractively",
			"Manage inventory and order kitchen supplies",
			"Keep the kitchen clean and follow hygiene rules",
			"Train junior kitchen staff",
			"Create new menu items and specials",
		},
		Skills: []string{
			"Cooking techniques (knife skills, grilling, sautéing, baking)",
			"Knowledge of Nepali and international cuisines",
			"Kitchen management and team leadership",
			"Menu planning and food cost management",
			"Food safety and hygiene standards",
			"Creativity and presentation skills",
			"Time management during busy service periods",
		},
		SalaryMin:   200000,
		SalaryMax:   1500000,
		Difficulty:  3,
		FutureProof: 68,
		EducationReq: "Diploma in Culinary Arts from institutes like NAAC, KATH, or Panchakanya. Practical experience is equally important. Many great chefs started as kitchen helpers and learned on the job.",
		Outlook:      "Nepal's restaurant and hotel industry is growing with tourism. New cafes and restaurants open every year in Kathmandu, Pokhara, and other tourist areas. Good chefs are always in demand. International opportunities are also available for Nepali chefs in India, the Middle East, and beyond.",
		Tags:         []string{"hospitality", "food", "creative", "kitchen", "hotel"},
		Resources: []resourceSeed{
			{Title: "NAAC Culinary Arts Program", URL: "https://www.naac.edu.np", Description: "Culinary arts training and certification in Nepal"},
			{Title: "Restaurant and Bar Association Nepal", URL: "https://www.ranepal.com", Description: "Resources and networking for food service professionals"},
			{Title: "Merojob Hospitality Jobs", URL: "https://www.merojob.com", Description: "Find chef and kitchen jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Start in any kitchen job to learn basics", Description: "Get a job as a kitchen helper (commis) or dishwasher in a restaurant or hotel kitchen. Watch how chefs work. Learn basic knife skills, food preparation, and kitchen hygiene. Work hard and show enthusiasm. Ask questions when the chef is not too busy. This is your first step. Even washing dishes teaches you about kitchen flow. Stay humble and keep learning.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Find kitchen jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Basic knife skills tutorial (YouTube)", URL: "https://www.youtube.com/results?search_query=basic+knife+skills+for+beginners"},
				{Title: "Kitchen hygiene standards", URL: "https://www.youtube.com/results?search_query=kitchen+hygiene+tips+nepal"},
			}},
			{StepNumber: 2, Title: "Get culinary training or apprenticeship", Description: "Enroll in a culinary arts program at NAAC, KATH, or a similar institute. If you cannot afford formal training, find a chef who will take you as an apprentice. Learn cooking techniques, food safety, and kitchen management. Practice at home — cook for your family and friends. The more you cook, the better you get. Watch free cooking tutorials online. Learn both Nepali dishes and international cuisines.", Duration: "1-2 years", Links: []roadmapLink{
				{Title: "NAAC Culinary Arts courses", URL: "https://www.naac.edu.np"},
				{Title: "Free cooking tutorials (YouTube)", URL: "https://www.youtube.com/results?search_query=learn+to+cook+for+beginners"},
				{Title: "Nepali cuisine cooking videos", URL: "https://www.youtube.com/results?search_query=nepali+food+recipes"},
			}},
			{StepNumber: 3, Title: "Work as a line cook and master a station", Description: "As a line cook, you will be responsible for one section of the kitchen: grill, fry, vegetable prep, or pastry. Master your station. Be fast, clean, and consistent. Learn to work during rush hours without getting stressed. Good line cooks are the backbone of every kitchen. After mastering one station, learn others. A chef who can work every station is very valuable.", Duration: "1-3 years", Links: []roadmapLink{
				{Title: "How to be a good line cook (YouTube)", URL: "https://www.youtube.com/results?search_query=line+cook+training"},
				{Title: "Restaurant and Bar Association Nepal - Training", URL: "https://www.ranepal.com"},
				{Title: "Professional cooking techniques", URL: "https://www.youtube.com/results?search_query=professional+cooking+techniques"},
			}},
			{StepNumber: 4, Title: "Become a sous chef and help run the kitchen", Description: "The sous chef is the second-in-command in the kitchen. You will manage staff, order supplies, control food costs, and help the head chef. Learn menu planning and pricing. Understand how to make a profit while maintaining quality. Learn to handle pressure and solve problems quickly. A good sous chef is ready to become a head chef.", Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Sous chef responsibilities (YouTube)", URL: "https://www.youtube.com/results?search_query=sous+chef+training"},
				{Title: "Food cost management for chefs", URL: "https://www.youtube.com/results?search_query=food+cost+percentage+chef"},
				{Title: "Kitchen leadership and management", URL: "https://www.youtube.com/results?search_query=kitchen+management+skills"},
			}},
			{StepNumber: 5, Title: "Become head chef or executive chef", Description: "As head chef, you run the entire kitchen. You create menus, manage the budget, hire staff, and ensure quality. In big hotels, the executive chef oversees multiple restaurants. This is a position of great responsibility and good pay. Continue learning — attend workshops, try new cuisines, travel for inspiration. The best chefs never stop learning.", Duration: "2-5 years", Links: []roadmapLink{
				{Title: "Executive chef career path", URL: "https://www.ranepal.com"},
				{Title: "Menu design and development (YouTube)", URL: "https://www.youtube.com/results?search_query=menu+planning+chef"},
				{Title: "International chef certification", URL: "https://www.youtube.com/results?search_query=chef+certification+nepal"},
			}},
			{StepNumber: 6, Title: "Open your own restaurant or become a celebrity chef", Description: "Many great chefs eventually open their own restaurant. Start small — a small cafe or eatery. Build a reputation for great food. Use social media to showcase your dishes. Participate in food festivals and competitions. Some Nepali chefs have become famous on YouTube and TV. Your cooking can make people happy and that is a wonderful thing. Food brings people together.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "How to start a restaurant in Nepal", URL: "https://www.ranepal.com"},
				{Title: "Restaurant business plan guide (YouTube)", URL: "https://www.youtube.com/results?search_query=start+restaurant+nepal"},
				{Title: "Building a chef brand on social media", URL: "https://www.youtube.com/results?search_query=become+famous+chef+nepal"},
			}},
		},
	}
}

func travelAgent() careerSeed {
	return careerSeed{
		CategoryName: "Tourism & Hospitality",
		CategorySlug: "tourism-hospitality",
		CategoryIcon: "✈️",
		Title:        "Travel Agent",
		Slug:         "travel-agent",
		Summary:      "Travel agents help people plan and book their trips. They arrange flights, hotels, tours, and other travel services for clients.",
		Description:  "A travel agent helps clients plan and book their travel. They arrange flights, train or bus tickets, hotel accommodation, tour packages, travel insurance, and visas. Travel agents work in travel agencies, or they can work independently from home. In Nepal, travel agents serve both outbound travelers (Nepalis going abroad) and inbound travelers (tourists coming to Nepal). The job requires good knowledge of destinations, airlines, hotels, and booking systems. Travel agents earn commission from bookings and sometimes charge service fees. The rise of online booking has changed the industry but many people still prefer using an agent they trust.",
		DailyTasks: []string{
			"Talk to clients about their travel plans and budget",
			"Search and compare flights, hotels, and tour options",
			"Book tickets, hotels, and tours for clients",
			"Help clients with visa applications and travel insurance",
			"Issue tickets and send confirmation documents",
			"Handle changes, cancellations, and refunds",
			"Keep up to date with travel deals and promotions",
		},
		Skills: []string{
			"Knowledge of travel destinations and booking systems",
			"Customer service and communication skills",
			"Computer skills and familiarity with booking software",
			"English language proficiency",
			"Sales and negotiation skills",
			"Attention to detail (dates, names, documents)",
			"Problem-solving when travel plans go wrong",
		},
		SalaryMin:   200000,
		SalaryMax:   800000,
		Difficulty:  2,
		FutureProof: 55,
		EducationReq: "SLC/SEE pass minimum. Diploma or certificate in Travel and Tourism from institutes like NAAC or KATH. IATA certification is highly valued for airline ticketing. Good computer skills essential.",
		Outlook:      "Online booking sites are competing with traditional travel agents, but agents who offer good service and advice still thrive. Many people prefer an agent for complex trips, group bookings, or when they need expert advice. Specializing in a niche (trekking, pilgrimage, business travel) helps.",
		Tags:         []string{"tourism", "travel", "customer-service", "office-job", "booking"},
		Resources: []resourceSeed{
			{Title: "NAAC Travel and Tourism Programs", URL: "https://www.naac.edu.np", Description: "Travel and tourism courses and certification in Nepal"},
			{Title: "IATA Training Nepal", URL: "https://www.iata.org/training", Description: "International airline ticketing and travel certification"},
			{Title: "Nepal Association of Travel Agents", URL: "https://www.natanepal.com", Description: "Resources and networking for travel agents in Nepal"},
			{Title: "Merojob Tourism Jobs", URL: "https://www.merojob.com", Description: "Find travel agent jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn about travel and destinations", Description: "Read about different countries, their cultures, visa requirements, and popular attractions. Learn about airlines, flight routes, and booking classes. Study maps and time zones. Travel yourself when you can — personal experience makes you a better advisor. Follow travel news and blogs. Understanding destinations is the foundation of being a good travel agent.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Nepal Association of Travel Agents - Resources", URL: "https://www.natanepal.com"},
				{Title: "Travel and tourism basics course (YouTube)", URL: "https://www.youtube.com/results?search_query=travel+agent+training+nepal"},
				{Title: "World travel destinations guide", URL: "https://www.youtube.com/results?search_query=travel+destinations+guide"},
			}},
			{StepNumber: 2, Title: "Get trained in booking systems and ticketing", Description: "Learn the Global Distribution Systems (GDS) like Amadeus, Galileo, or Sabre — these are used to book flights worldwide. Get IATA certification for airline ticketing. Learn hotel booking platforms. These skills are essential and make you employable. Many travel institutes in Kathmandu offer these courses. Computer skills are a must — practice typing fast and accurately.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "IATA training and certification", URL: "https://www.iata.org/training"},
				{Title: "Amadeus booking system tutorial (YouTube)", URL: "https://www.youtube.com/results?search_query=amadeus+training+for+beginners"},
				{Title: "NAAC - Travel and tourism courses", URL: "https://www.naac.edu.np"},
			}},
			{StepNumber: 3, Title: "Work in a travel agency to gain experience", Description: "Start as a junior travel consultant or reservation agent in a travel agency. Learn how to handle client inquiries, make bookings, issue tickets, and manage files. Watch how senior agents handle difficult clients and solve problems. Build your speed and accuracy. A single wrong letter in a name can cause big problems — attention to detail is everything in this job.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Find travel agency jobs in Nepal", URL: "https://www.merojob.com"},
				{Title: "NATA Nepal - Member travel agencies", URL: "https://www.natanepal.com"},
				{Title: "Travel agent customer service skills (YouTube)", URL: "https://www.youtube.com/results?search_query=travel+agent+customer+service"},
			}},
			{StepNumber: 4, Title: "Build a network of regular clients", Description: "Happy clients come back and tell their friends. Give every client your best service. Remember their preferences. Follow up after their trip. Build a database of client contacts. Offer special deals to repeat customers. A travel agent with a loyal client base has a stable income. Word of mouth is powerful in this business. Be the agent people trust with their holidays.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Client relationship tips for travel agents", URL: "https://www.youtube.com/results?search_query=travel+agent+client+management"},
				{Title: "NATA - Business development resources", URL: "https://www.natanepal.com"},
				{Title: "Using social media as a travel agent", URL: "https://www.youtube.com/results?search_query=social+media+travel+agent+nepal"},
			}},
			{StepNumber: 5, Title: "Specialize in a type of travel", Description: "General travel agents earn less than specialists. Specialize in a niche: trekking and adventure tours, pilgrimage travel (Muktinath, Pashupatinath, Mansarovar), honeymoon packages, business travel, or student travel. Become the expert in your niche. Know every detail. Specialists can charge higher service fees and get better commissions.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Niche travel markets in Nepal", URL: "https://www.tourismboard.gov.np"},
				{Title: "Pilgrimage tour packages - Guide", URL: "https://www.welcomenepal.com"},
				{Title: "Merojob - Specialized travel jobs", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 6, Title: "Start your own travel agency or go independent", Description: "With experience and a client base, you can start your own travel agency. Register with the Nepal Tourism Board and NATA. Start small — a home office is fine. Build a website. Focus on your niche. The internet makes it possible for one-person agencies to reach clients worldwide. Being your own boss is hard work but very rewarding. You have the power to make people's dream trips come true.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a travel agency in Nepal", URL: "https://www.natanepal.com"},
				{Title: "Home-based travel agency guide (YouTube)", URL: "https://www.youtube.com/results?search_query=start+travel+agency+from+home+nepal"},
				{Title: "Nepal Tourism Board - Agency licensing", URL: "https://www.tourismboard.gov.np"},
			}},
		},
	}
}

func bartender() careerSeed {
	return careerSeed{
		CategoryName: "Tourism & Hospitality",
		CategorySlug: "tourism-hospitality",
		CategoryIcon: "🍸",
		Title:        "Bartender / Barista",
		Slug:         "bartender",
		Summary:      "Bartenders mix and serve drinks in hotels, bars, restaurants, and cafes. They create cocktails, serve coffee, and give customers a great experience.",
		Description:  "A bartender or barista serves drinks to customers. Bartenders mix cocktails, pour beer and wine, and serve spirits. Baristas specialize in coffee drinks like espresso, cappuccino, and latte. In Nepal, bartenders and baristas work in hotels, restaurants, bars, clubs, and cafes. The coffee shop culture is growing fast in Kathmandu, Pokhara, and other cities. Many bartenders and baristas learn on the job, but formal training helps you get better jobs. The job requires good customer service, speed, knowledge of drinks, and the ability to work evenings and weekends. Tips can be a significant part of income.",
		DailyTasks: []string{
			"Greet customers and take drink orders",
			"Mix and serve alcoholic and non-alcoholic drinks",
			"Make espresso-based coffee drinks (for baristas)",
			"Keep the bar area clean and organized",
			"Check IDs to make sure customers are legal age",
			"Handle cash and process payments",
			"Stock supplies — liquor, glassware, napkins, garnishes",
		},
		Skills: []string{
			"Knowledge of cocktails, beer, wine, and spirits",
			"Coffee making skills (espresso, latte art) for baristas",
			"Customer service and friendly attitude",
			"Speed and efficiency during busy times",
			"Cash handling and basic math",
			"Memory for drink recipes and customer orders",
			"Responsibility with alcohol service",
		},
		SalaryMin:   150000,
		SalaryMax:   600000,
		Difficulty:  2,
		FutureProof: 60,
		EducationReq: "SLC/SEE pass minimum. Bartending or barista training from institutes like NTB/NAAC hospitality programs. Many learn on the job. International certifications like WSET (wine) or SCA (coffee) help.",
		Outlook:      "Nepal's cafe and bar scene is growing, especially in tourist areas. New cocktail bars and specialty coffee shops open regularly. Good bartenders and baristas are always in demand. International opportunities in the Middle East, India, and cruise ships are available for experienced professionals.",
		Tags:         []string{"hospitality", "service", "nightlife", "coffee", "tips"},
		Resources: []resourceSeed{
			{Title: "Barista Training Nepal", URL: "https://www.youtube.com/results?search_query=barista+training+nepal", Description: "Free barista training videos and coffee making tutorials"},
			{Title: "Restaurant and Bar Association Nepal", URL: "https://www.ranepal.com", Description: "Resources for bar and restaurant professionals in Nepal"},
			{Title: "Merojob Hospitality Jobs", URL: "https://www.merojob.com", Description: "Find bartender and barista jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Get a job as a bar back or assistant", Description: "Start as a bar back (helper) or assistant in a bar or cafe. You will clean glasses, restock supplies, and help the bartender. Watch how drinks are made. Learn the names of liquors and cocktail ingredients. Ask questions. Practice making simple drinks when it is slow. This is your training ground. Work hard and be reliable — bartenders notice and will teach you.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Find bar jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Bar tending basics for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=bartending+for+beginners"},
				{Title: "Restaurant and Bar Association Nepal", URL: "https://www.ranepal.com"},
			}},
			{StepNumber: 2, Title: "Learn drink recipes and techniques", Description: "Memorize the most popular cocktail recipes: margarita, mojito, old fashioned, martini, and local favorites. Learn proper pouring technique, how to use a jigger, how to shake and stir drinks. For baristas, learn espresso extraction, milk steaming, and latte art. Practice at home with basic ingredients. Watch online tutorials. Speed comes with practice. Accuracy matters more than speed at first.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Cocktail recipes for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=cocktail+recipes+for+beginners"},
				{Title: "Latte art tutorial for baristas (YouTube)", URL: "https://www.youtube.com/results?search_query=latte+art+for+beginners"},
				{Title: "Specialty Coffee Association - Training", URL: "https://www.sca.coffee"},
			}},
			{StepNumber: 3, Title: "Get certified in bartending or coffee", Description: "Take a bartending course from a recognized institute. Learn about responsible alcohol service, drink recipes, and bar management. For baristas, get SCA (Specialty Coffee Association) certification. These certifications help you get jobs in better establishments and command higher pay. Many training centers in Kathmandu offer these courses at reasonable prices.", Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Bartending course in Kathmandu", URL: "https://www.ranepal.com"},
				{Title: "SCA Coffee Skills Program", URL: "https://www.sca.coffee"},
				{Title: "Responsible alcohol service training", URL: "https://www.youtube.com/results?search_query=responsible+beverage+service+training"},
			}},
			{StepNumber: 4, Title: "Work in busy bars to build speed and experience", Description: "Work in a busy hotel bar, nightclub, or high-volume restaurant. Speed, accuracy, and grace under pressure come from experience. Learn to handle multiple orders at once. Develop your own style and signature drinks. Build regular customers who ask for you by name. A bartender with a following can work anywhere. In the service industry, your personality is part of what people pay for.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Fine dining bartending tips (YouTube)", URL: "https://www.youtube.com/results?search_query=hotel+bartending+tips"},
				{Title: "Up-selling techniques for bartenders", URL: "https://www.youtube.com/results?search_query=bartender+upselling+techniques"},
				{Title: "Hotel bar jobs in Nepal", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 5, Title: "Learn about wine, spirits, and premium products", Description: "Study wine regions, grape varieties, and food pairing. Learn about whiskey, rum, vodka, tequila, and premium spirits. Get WSET (Wine and Spirit Education Trust) certification if possible. Knowledge of premium drinks impresses customers and helps you work in upscale venues. High-end hotels and restaurants pay much more and their customers expect expert guidance.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "WSET wine certification program", URL: "https://www.wsetglobal.com"},
				{Title: "Whiskey knowledge for bartenders (YouTube)", URL: "https://www.youtube.com/results?search_query=whiskey+guide+bartenders"},
				{Title: "Wine pairing basics (YouTube)", URL: "https://www.youtube.com/results?search_query=wine+pairing+for+beginners"},
			}},
			{StepNumber: 6, Title: "Become a bar manager or open your own bar", Description: "With 3-5 years of experience, become a bar manager. Manage inventory, create cocktail menus, train staff, control costs, and ensure profitability. Some bartenders go on to open their own bars or cafes. Nepal's nightlife and cafe culture are growing. With the right concept and location, a bar or cafe can be a successful business. The best part? You get to create a place where people make happy memories.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Bar management guide (YouTube)", URL: "https://www.youtube.com/results?search_query=bar+management+tips"},
				{Title: "How to open a bar in Nepal", URL: "https://www.ranepal.com"},
				{Title: "Nepal alcohol licensing requirements", URL: "https://www.merojob.com"},
			}},
		},
	}
}

func flightAttendant() careerSeed {
	return careerSeed{
		CategoryName: "Tourism & Hospitality",
		CategorySlug: "tourism-hospitality",
		CategoryIcon: "👩‍✈️",
		Title:        "Flight Attendant",
		Slug:         "flight-attendant",
		Summary:      "Flight attendants ensure passenger safety and comfort on airplanes. They serve food, demonstrate safety procedures, and handle emergencies on flights.",
		Description:  "A flight attendant (also called cabin crew) takes care of passengers on airplanes. They greet passengers, help with luggage, demonstrate safety procedures, serve meals and drinks, and handle any emergencies that occur during the flight. Flight attendants in Nepal work for airlines like Nepal Airlines, Buddha Air, Yeti Airlines, and international airlines based in Nepal. The job requires excellent customer service, language skills, physical fitness, and the ability to stay calm in emergencies. Flight attendants spend many nights away from home but get to travel and meet people from all over the world. Height and appearance requirements vary by airline.",
		DailyTasks: []string{
			"Greet passengers as they board the plane",
			"Demonstrate safety procedures before takeoff",
			"Serve meals, drinks, and snacks during the flight",
			"Check seat belts, tray tables, and luggage storage",
			"Monitor cabin for any unusual situations",
			"Handle passenger requests and complaints",
			"Respond to medical or safety emergencies",
		},
		Skills: []string{
			"Customer service and hospitality skills",
			"English language (other languages a plus)",
			"First aid and emergency response training",
			"Public speaking (making announcements)",
			"Calm under pressure and quick thinking",
			"Physical fitness and stamina",
			"Teamwork with other crew members",
		},
		SalaryMin:   400000,
		SalaryMax:   1500000,
		Difficulty:  3,
		FutureProof: 50,
		EducationReq: "SLC/SEE pass minimum, PCL/+2 preferred. Cabin crew training from aviation training institutes like Civil Aviation Authority of Nepal (CAAN) approved schools. Height requirements typically 157-160cm+ for females, 170cm+ for males.",
		Outlook:      "Nepal's aviation industry is growing with new airlines and routes. However, competition for flight attendant jobs is very high. International airlines sometimes recruit from Nepal. The job has good pay and benefits but irregular hours and time away from family. Automation may reduce some cabin crew roles but safety requirements keep humans essential.",
		Tags:         []string{"aviation", "travel", "hospitality", "customer-service", "adventure"},
		Resources: []resourceSeed{
			{Title: "Civil Aviation Authority Nepal", URL: "https://www.caanepal.gov.np", Description: "Official aviation regulatory body — training and certification info"},
			{Title: "Nepal Airlines Corporation", URL: "https://www.nepalairlines.com.np", Description: "Career opportunities with Nepal's national airline"},
			{Title: "Merojob Aviation Jobs", URL: "https://www.merojob.com", Description: "Find flight attendant and aviation jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Meet the basic requirements and apply", Description: "Check the requirements for Nepali airlines: minimum height, weight proportional, good vision (with or without glasses), clear skin, and no visible tattoos. Good English is essential. Start working on your communication and presentation skills. Practice speaking clearly and confidently. Research different airlines and their requirements. Each airline has its own standards and preferences.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "CAAN - Aviation career requirements", URL: "https://www.caanepal.gov.np"},
				{Title: "How to become a flight attendant (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+become+flight+attendant+nepal"},
				{Title: "English communication skills practice", URL: "https://www.bbc.co.uk/learningenglish"},
			}},
			{StepNumber: 2, Title: "Get cabin crew training", Description: "Enroll in a CAAN-approved cabin crew training program. These programs cover safety procedures, first aid, emergency evacuation, firefighting, water survival, customer service, and grooming. Training takes 3-6 months. Some airlines have their own training academy. A certificate from a recognized school increases your chances of being hired. This is an investment in your career.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "CAAN - Approved training organizations", URL: "https://www.caanepal.gov.np"},
				{Title: "Cabin crew training in Kathmandu (YouTube)", URL: "https://www.youtube.com/results?search_query=cabin+crew+training+nepal"},
				{Title: "Aviation training institutes in Nepal", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 3, Title: "Learn first aid and emergency response", Description: "Flight attendants are trained first responders. Get certified in first aid, CPR, and AED use. Learn about medical emergencies that can happen on flights: allergic reactions, heart attacks, panic attacks, childbirth. Practice emergency procedures: evacuating a plane in 90 seconds, using fire extinguishers, deploying life rafts. This training saves lives. Take it seriously.", Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Red Cross First Aid certification Nepal", URL: "https://www.nrcs.org"},
				{Title: "Aviation first aid training (YouTube)", URL: "https://www.youtube.com/results?search_query=aviation+first+aid+training"},
				{Title: "CPR and AED certification (free)", URL: "https://www.youtube.com/results?search_query=cpr+training+for+beginners"},
			}},
			{StepNumber: 4, Title: "Apply to airlines and practice for interviews", Description: "Prepare your resume and professional photo. Practice interview questions: Why do you want to be a flight attendant? How do you handle difficult passengers? Why should we hire you? Airlines look for candidates who are friendly, professional, confident, and well-groomed. Be polite and smiling throughout the interview process. Group interviews test your teamwork. Role-play scenarios test your problem-solving under pressure.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Flight attendant interview tips (YouTube)", URL: "https://www.youtube.com/results?search_query=flight+attendant+interview+nepal"},
				{Title: "Airline recruitment in Nepal", URL: "https://www.merojob.com"},
				{Title: "Grooming and presentation for cabin crew", URL: "https://www.youtube.com/results?search_query=cabin+crew+grooming+standards"},
			}},
			{StepNumber: 5, Title: "Pass training and start flying", Description: "Once hired, you will go through the airline's own training program. This is intense and you must pass all exams. Learn the specific procedures for your airline's aircraft type. Learn service standards, meal service procedures, and in-flight sales. Your first few flights will be nerve-wracking but everyone starts somewhere. Experienced crew will guide you. Ask questions. Be helpful. The first year is the hardest but it gets easier.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Your first year as cabin crew (YouTube)", URL: "https://www.youtube.com/results?search_query=first+year+flight+attendant"},
				{Title: "Airline safety and service standards", URL: "https://www.caanepal.gov.np"},
				{Title: "Nepali airline crew experience", URL: "https://www.youtube.com/results?search_query=nepal+airlines+cabin+crew"},
			}},
			{StepNumber: 6, Title: "Advance to senior crew or international airlines", Description: "After 2-3 years, you can become a senior flight attendant or purser (in charge of all cabin crew). International airlines based in the Gulf (Emirates, Qatar, Etihad) recruit Nepali cabin crew. The pay and benefits are much better. Apply to international airlines when you have experience. Some crew move into ground roles like training, recruitment, or management. The sky is literally the limit.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "International airline recruitment Nepali crew", URL: "https://www.merojob.com"},
				{Title: "Emirates cabin crew careers", URL: "https://www.emiratesgroupcareers.com"},
				{Title: "Senior cabin crew career progression (YouTube)", URL: "https://www.youtube.com/results?search_query=senior+cabin+crew+promotion"},
			}},
		},
	}
}

func raftingGuide() careerSeed {
	return careerSeed{
		CategoryName: "Tourism & Hospitality",
		CategorySlug: "tourism-hospitality",
		CategoryIcon: "🛶",
		Title:        "Rafting Guide",
		Slug:         "rafting-guide",
		Summary:      "Rafting guides lead groups of adventurers on river rafting trips through Nepal's exciting white-water rivers.",
		Description:  "A rafting guide leads groups on white-water rafting trips. Nepal has some of the best rivers in the world for rafting and kayaking — Trishuli, Bhote Koshi, Kali Gandaki, Karnali, and Sun Koshi. Guides are responsible for the safety of their group, navigating rapids, and ensuring everyone has fun. The job requires swimming skills, knowledge of river safety, first aid certification, and a rafting guide license from the Nepal Tourism Board. Rafting guides work during the main seasons (September-November and March-May). It is a physically active, outdoor job that attracts adventure-loving people. Many guides also work as kayakers or canyoning guides.",
		DailyTasks: []string{
			"Inspect and prepare rafting equipment (rafts, paddles, life jackets, helmets)",
			"Brief clients on safety procedures and paddling commands",
			"Guide the raft through rapids and calm sections",
			"Monitor clients for safety and enjoyment",
			"Manage rescue situations if someone falls in the water",
			"Set up camp and cook meals on multi-day trips",
			"Maintain and repair equipment after trips",
		},
		Skills: []string{
			"White-water rafting and paddling techniques",
			"Swift water rescue and safety management",
			"First aid and CPR certification",
			"Knowledge of Nepali rivers and rapid classifications",
			"Swimming and physical fitness",
			"Customer service and group management",
			"Equipment maintenance and repair",
		},
		SalaryMin:   200000,
		SalaryMax:   800000,
		Difficulty:  3,
		FutureProof: 62,
		EducationReq: "SLC/SEE pass minimum. Must complete Nepal Tourism Board rafting guide training and pass license exam. First aid and swift water rescue certification required. Swimming ability essential.",
		Outlook:      "River rafting is a popular adventure activity in Nepal. The number of rafters is growing each year. The government is promoting adventure tourism. Experienced guides can work in other countries too. Seasonality is the main challenge — work is limited to 6-8 months per year.",
		Tags:         []string{"tourism", "adventure", "outdoor", "rivers", "fitness"},
		Resources: []resourceSeed{
			{Title: "Nepal Tourism Board - Rafting Guide License", URL: "https://www.tourismboard.gov.np", Description: "Official rafting guide licensing requirements and training"},
			{Title: "Nepal Association of Rafting Agencies", URL: "https://www.raftingassociation.org.np", Description: "Professional body for rafting guides and agencies in Nepal"},
			{Title: "Merojob Tourism Jobs", URL: "https://www.merojob.com", Description: "Find rafting guide and adventure tourism jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn to swim well and overcome fear of water", Description: "If you cannot swim, learn immediately. Take swimming lessons in a pool or river. Being comfortable in moving water is essential. Practice floating, treading water, and swimming in currents. Learn to read water — where the currents are, where the rocks are, where the safe channels are. Spend time on and in the river. A good rafting guide is first a good swimmer.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Swimming lessons for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=learn+to+swim+for+beginners"},
				{Title: "River safety and water reading skills", URL: "https://www.youtube.com/results?search_query=how+to+read+a+river"},
				{Title: "Nepal Association of Rafting Agencies - Training", URL: "https://www.raftingassociation.org.np"},
			}},
			{StepNumber: 2, Title: "Go rafting as a guest to experience it", Description: "Go on several rafting trips as a paying guest or helper. Experience different rivers — Trishuli (class 2-3, good for beginners), Bhote Koshi (class 4-5, exciting), Seti (scenic), and Karnali (remote and wild). Watch how guides handle the group, read rapids, and manage safety. Ask questions. Learn the paddling commands. Feel what it is like to be in the raft when it drops into a big rapid. You need to love this feeling to be a rafting guide.", Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Best rafting rivers in Nepal", URL: "https://www.welcomenepal.com"},
				{Title: "Nepal Association of Rafting Agencies - Rivers", URL: "https://www.raftingassociation.org.np"},
				{Title: "Rafting experiences and videos (YouTube)", URL: "https://www.youtube.com/results?search_query=rafting+nepal+experience"},
			}},
			{StepNumber: 3, Title: "Complete your guide training and certification", Description: "Enroll in a rafting guide training program approved by the Nepal Tourism Board. The course covers rafting techniques, river safety, rescue skills, first aid, group management, and environmental practices. Pass the written and practical exams to get your rafting guide license. Also get a Swift Water Rescue certification and Wilderness First Aid. These certifications are required by law and by insurance companies.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Nepal Tourism Board - Guide training info", URL: "https://www.tourismboard.gov.np"},
				{Title: "Swift water rescue certification", URL: "https://www.youtube.com/results?search_query=swift+water+rescue+training"},
				{Title: "Rafting guide course (YouTube)", URL: "https://www.youtube.com/results?search_query=rafting+guide+course+nepal"},
			}},
			{StepNumber: 4, Title: "Work with an established rafting company", Description: "Join a rafting company as an assistant guide or junior guide. Start with easier rivers (class 2-3). Learn company procedures, client service standards, and logistics management. Build your experience on different rivers and with different types of groups (families, adventure seekers, corporate groups, school trips). Each group needs a different approach. Listen to feedback and improve your skills every trip.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Find rafting guide jobs in Nepal", URL: "https://www.merojob.com"},
				{Title: "NARA - Member rafting agencies", URL: "https://www.raftingassociation.org.np"},
				{Title: "How to lead rafting trips (YouTube)", URL: "https://www.youtube.com/results?search_query=leading+rafting+trips+tips"},
			}},
			{StepNumber: 5, Title: "Learn advanced skills — kayaking, rescue, and first aid", Description: "Learn to kayak — kayakers can guide from their kayak alongside rafts. Take advanced swift water rescue courses. Become a certified Wilderness First Responder. Learn to read advanced rapids (class 4-5). The more skills you have, the more valuable you are. Advanced guides can lead multi-day expeditions on remote rivers like the Karnali and Sun Koshi. These trips pay better and are more adventurous.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Kayaking lessons for rafting guides (YouTube)", URL: "https://www.youtube.com/results?search_query=kayak+training+for+rafting"},
				{Title: "Advanced swift water rescue techniques", URL: "https://www.youtube.com/results?search_query=advanced+swift+water+rescue"},
				{Title: "Wilderness First Responder certification", URL: "https://www.nrcs.org"},
			}},
			{StepNumber: 6, Title: "Become a senior guide or start your own company", Description: "With 3-5 years of experience, become a senior guide leading the most challenging trips. Some guides become trainers for new guides. Others start their own rafting company. The Nepal Association of Rafting Agencies can help with business setup. Combine rafting with other activities like canyoning, bungee jumping, or trekking. Nepal's rivers are world-class and people come from all over the world to experience them. Being a rafting guide is more than a job — it is a lifestyle.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a rafting company in Nepal", URL: "https://www.raftingassociation.org.np"},
				{Title: "Nepal adventure tourism opportunities", URL: "https://www.tourismboard.gov.np"},
				{Title: "Senior rafting guide career tips (YouTube)", URL: "https://www.youtube.com/results?search_query=rafting+guide+career+nepal"},
			}},
		},
	}
}

func itSupport() careerSeed {
	return careerSeed{
		CategoryName: "Technology & IT",
		CategorySlug: "technology-it",
		CategoryIcon: "🖥️",
		Title:        "IT Support Specialist",
		Slug:         "it-support",
		Summary:      "IT support specialists help people and companies fix computer problems. They set up systems, install software, and keep everything running smoothly.",
		Description:  "An IT support specialist helps people use technology. When a computer breaks, a printer stops working, or the internet goes down, the IT support person fixes it. They set up new computers, install software, create user accounts, and make sure the network is secure. In Nepal, IT support jobs are growing as more companies and organizations use computers and the internet. IT support specialists work in schools, hospitals, banks, government offices, and technology companies. You can start with basic computer knowledge and learn more on the job. Certifications like CompTIA A+ and ITIL are valuable but not always required.",
		DailyTasks: []string{
			"Respond to help desk tickets and user requests",
			"Fix computer hardware and software problems",
			"Install and update software on company computers",
			"Set up user accounts and email systems",
			"Maintain computer networks and servers",
			"Run backups and make sure data is safe",
			"Train users on new technology and best practices",
		},
		Skills: []string{
			"Computer hardware troubleshooting and repair",
			"Windows, Linux, or macOS operating systems",
			"Network basics (TCP/IP, DNS, routers, switches)",
			"Customer service and patience with non-technical users",
			"Problem-solving and logical thinking",
			"Basic cybersecurity awareness",
			"Documentation and ticket tracking",
		},
		SalaryMin:   250000,
		SalaryMax:   900000,
		Difficulty:  3,
		FutureProof: 68,
		EducationReq: "SLC/SEE pass plus computer training from institutes like Broadway Infosys, Aptech, or similar. PCL/+2 in Computer Science helpful. CompTIA A+ certification valuable. Practical skills matter more than degrees.",
		Outlook:      "Every company needs IT support. As Nepal's economy digitizes, demand for IT support is growing. Government offices, banks, hospitals, and schools all need IT staff. The work can be entry-level but leads to higher IT roles. Outsourcing and remote support are creating new opportunities.",
		Tags:         []string{"tech", "support", "help-desk", "computers", "entry-level"},
		Resources: []resourceSeed{
			{Title: "Broadway Infosys - IT Training", URL: "https://www.broadwayinfosys.com", Description: "Leading IT training institute in Nepal with support courses"},
			{Title: "CompTIA A+ Certification", URL: "https://www.comptia.org/certifications/a", Description: "Industry standard certification for IT support professionals"},
			{Title: "Merojob IT Jobs", URL: "https://www.merojob.com", Description: "Find IT support and help desk jobs in Nepal"},
			{Title: "IT Support Free Course (Google)", URL: "https://www.coursera.org/professional-certificates/google-it-support", Description: "Free IT support training from Google on Coursera"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Build your computer knowledge from the ground up", Description: "Learn how computers work: hardware parts (CPU, RAM, hard drive, motherboard), operating systems (Windows is most common in Nepal), and common software. Practice building a computer from parts. Install Windows from scratch. Learn to troubleshoot common problems: slow computer, no internet, blue screen errors. Free online resources can teach you all of this. Help friends and family with their computer problems — hands-on practice is the best teacher.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Google IT Support free course", URL: "https://www.coursera.org/professional-certificates/google-it-support"},
				{Title: "Computer hardware basics (YouTube)", URL: "https://www.youtube.com/results?search_query=computer+hardware+basics+for+beginners"},
				{Title: "Windows troubleshooting guide (YouTube)", URL: "https://www.youtube.com/results?search_query=windows+troubleshooting+basics"},
			}},
			{StepNumber: 2, Title: "Learn networking basics", Description: "Understand how computers connect to each other and the internet. Learn about IP addresses, routers, switches, DNS, DHCP, and Wi-Fi. Set up a small home network. Learn to configure a router. Understand the difference between LAN and WAN. Networking is a big part of IT support. You do not need to be a network engineer, but you need to understand the basics. Free online courses cover all of this.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Networking basics free course", URL: "https://www.youtube.com/results?search_query=networking+for+beginners"},
				{Title: "Cisco Networking Academy free courses", URL: "https://www.netacad.com"},
				{Title: "Home network setup tutorial (YouTube)", URL: "https://www.youtube.com/results?search_query=home+network+setup+guide"},
			}},
			{StepNumber: 3, Title: "Get a CompTIA A+ certification", Description: "CompTIA A+ is the industry standard certification for IT support. It covers hardware, networking, mobile devices, operating systems, security, and troubleshooting. Study for 2-3 months using free resources. The exam costs about $246 but it is worth it — certified professionals earn more and get better jobs. Many IT companies in Nepal recognize this certification. If you cannot afford the exam, study the material anyway — the knowledge is what matters.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "CompTIA A+ exam guide and resources", URL: "https://www.comptia.org/certifications/a"},
				{Title: "Free CompTIA A+ study materials (Professor Messer)", URL: "https://www.professormesser.com"},
				{Title: "IT training institutes in Nepal", URL: "https://www.broadwayinfosys.com"},
			}},
			{StepNumber: 4, Title: "Get your first IT support job", Description: "Apply for help desk or IT support roles in companies, schools, or government offices in Nepal. Update your resume to highlight your troubleshooting skills and any certifications. Be honest about what you know and eager to learn what you do not. Your first job may not pay much but the experience is invaluable. Solve every problem like it matters. Build a reputation for being reliable and helpful. Many IT managers started at the help desk.", Duration: "1-6 months", Links: []roadmapLink{
				{Title: "Find IT support jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "IT help desk resume tips (YouTube)", URL: "https://www.youtube.com/results?search_query=it+help+desk+resume+tips"},
				{Title: "Common IT support interview questions", URL: "https://www.youtube.com/results?search_query=it+support+interview+questions"},
			}},
			{StepNumber: 5, Title: "Learn advanced skills — servers, cloud, and security", Description: "Learn about Windows Server, Linux (Ubuntu), cloud platforms (AWS, Google Cloud), and cybersecurity basics. Learn how to manage users, set up file servers, and configure email systems. These skills help you move from basic support to system administration. System administrators earn much more than help desk staff. The IT field is all about continuous learning — keep studying even after you get the job.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Windows Server basics (YouTube)", URL: "https://www.youtube.com/results?search_query=windows+server+for+beginners"},
				{Title: "AWS Cloud Practitioner free course", URL: "https://aws.amazon.com/training/digital/aws-cloud-practitioner-essentials"},
				{Title: "Cybersecurity basics for IT support", URL: "https://www.youtube.com/results?search_query=cybersecurity+basics+for+beginners"},
			}},
			{StepNumber: 6, Title: "Specialize or move into management", Description: "After 2-3 years, choose a path: become a system administrator, network administrator, IT security specialist, or IT manager. Or start your own IT support business helping small companies in your area. Many small businesses in Nepal cannot afford full-time IT staff and would pay for reliable support. The IT field in Nepal is growing and good professionals are always needed. Keep learning and helping people with technology.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "IT career paths in Nepal", URL: "https://www.merojob.com"},
				{Title: "Starting an IT support business (YouTube)", URL: "https://www.youtube.com/results?search_query=start+it+support+business"},
				{Title: "System administration career guide", URL: "https://www.youtube.com/results?search_query=system+admin+career+nepal"},
			}},
		},
	}
}

func digitalMarketer() careerSeed {
	return careerSeed{
		CategoryName: "Technology & IT",
		CategorySlug: "technology-it",
		CategoryIcon: "📱",
		Title:        "Digital Marketer",
		Slug:         "digital-marketer",
		Summary:      "Digital marketers promote businesses on the internet using social media, email, search engines, and online ads to reach customers.",
		Description:  "A digital marketer helps businesses reach customers online. They manage social media accounts (Facebook, Instagram, TikTok), run online ads (Google Ads, Facebook Ads), send email newsletters, create content (videos, blogs, images), and track results using analytics. In Nepal, digital marketing is a fast-growing field because more businesses are moving online. Every company, from hotels to restaurants to clothing stores, needs digital marketing. You can work as a digital marketer for a company, at a marketing agency, or as a freelancer. The field changes fast but offers good pay and creative work.",
		DailyTasks: []string{
			"Create social media posts for Facebook, Instagram, and TikTok",
			"Monitor and respond to comments and messages",
			"Run and optimize online advertising campaigns",
			"Write blog posts, emails, and website content",
			"Analyze website traffic and campaign performance",
			"Research trending topics and keywords",
			"Create simple graphics and videos for posts",
		},
		Skills: []string{
			"Social media management (Facebook, Instagram, TikTok)",
			"Content creation (writing, basic graphic design, video)",
			"Search Engine Optimization (SEO) basics",
			"Online advertising (Google Ads, Facebook Ads)",
			"Analytics (Google Analytics, Meta Business Suite)",
			"Email marketing tools (Mailchimp, SendGrid)",
			"English and Nepali content writing",
		},
		SalaryMin:   250000,
		SalaryMax:   1200000,
		Difficulty:  3,
		FutureProof: 78,
		EducationReq: "SLC/SEE pass minimum. PCL/+2 or Bachelor's in Business/IT helpful. Digital marketing certifications from Google, Meta, and HubSpot are highly valued. Practical skills matter most — a good portfolio beats a degree.",
		Outlook:      "Digital marketing is growing very fast in Nepal. More businesses want an online presence. Social media usage in Nepal is among the highest in South Asia. The field is competitive but skilled digital marketers are in high demand. Remote work opportunities with international clients are growing.",
		Tags:         []string{"tech", "marketing", "social-media", "creative", "remote-work"},
		Resources: []resourceSeed{
			{Title: "Google Digital Marketing Certificate", URL: "https://skillshop.withgoogle.com", Description: "Free digital marketing certification from Google"},
			{Title: "HubSpot Academy", URL: "https://academy.hubspot.com", Description: "Free marketing courses and certifications"},
			{Title: "Facebook Blueprint (Meta)", URL: "https://www.facebook.com/business/learn", Description: "Free social media marketing training from Meta"},
			{Title: "Merojob Marketing Jobs", URL: "https://www.merojob.com", Description: "Find digital marketing jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn how social media and online marketing work", Description: "Create accounts on Facebook, Instagram, TikTok, and LinkedIn if you have not already. Study how businesses use these platforms. Look at ads you see — why are they showing you this ad? Read free marketing blogs and watch YouTube tutorials. Learn the basic concepts: reach, engagement, conversion, click-through rate, ROI. Practice by creating a page for a hobby or your family business. Try posting regularly and see what gets attention.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Free digital marketing course (Google)", URL: "https://skillshop.withgoogle.com"},
				{Title: "Social media marketing basics (YouTube)", URL: "https://www.youtube.com/results?search_query=social+media+marketing+for+beginners"},
				{Title: "HubSpot Academy - Free marketing courses", URL: "https://academy.hubspot.com"},
			}},
			{StepNumber: 2, Title: "Learn content creation — writing, design, video", Description: "Good content is the heart of digital marketing. Learn to write engaging posts and articles. Learn basic graphic design using Canva or Photoshop. Learn to shoot and edit simple videos for social media. Practice every day. Create content for a fake brand or help a local business for free. Your portfolio matters more than your resume in marketing. Show what you can create, not just what you know.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Canva Design School (free)", URL: "https://www.canva.com/designschool/"},
				{Title: "Content writing tips for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=content+writing+for+beginners"},
				{Title: "Video editing for social media (YouTube)", URL: "https://www.youtube.com/results?search_query=social+media+video+editing+tips"},
			}},
			{StepNumber: 3, Title: "Get certified in digital marketing tools", Description: "Get free certifications from Google (Google Digital Marketing and E-commerce), Meta (Facebook Blueprint), and HubSpot. These are recognized by employers worldwide and show you have real skills. Learn Google Analytics to understand website traffic. Learn Google Ads to run search and display ads. Learn Facebook Ads Manager. Each certification takes 1-2 months of study. They are free or very low cost. Certified marketers get hired faster and earn more.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Google Digital Marketing Certification", URL: "https://skillshop.withgoogle.com"},
				{Title: "Facebook Blueprint Certification", URL: "https://www.facebook.com/business/learn"},
				{Title: "HubSpot Inbound Marketing Certification", URL: "https://academy.hubspot.com"},
			}},
			{StepNumber: 4, Title: "Get your first marketing job or freelance client", Description: "Apply to digital marketing roles in Nepal on Merojob, LinkedIn, and Facebook groups. Or start freelancing on platforms like Upwork, Fiverr, or directly approach small businesses in your area. Offer to manage their social media or run ads for a small fee. Your first client might pay very little. Do great work anyway. Build case studies. Get testimonials. Each successful campaign leads to the next client. The marketing field rewards results, not credentials.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Find marketing jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Freelance marketing on Upwork (YouTube)", URL: "https://www.youtube.com/results?search_query=freelance+marketing+nepal"},
				{Title: "Facebook groups for Nepali marketers", URL: "https://www.facebook.com"},
			}},
			{StepNumber: 5, Title: "Learn advanced skills — SEO, analytics, strategy", Description: "Learn Search Engine Optimization (SEO) to help websites rank on Google. Learn advanced analytics to measure campaign success. Understand marketing strategy: how to plan campaigns, set budgets, target audiences, and measure ROI. Learn marketing automation tools. The difference between a social media manager and a digital marketing strategist is big. Strategists plan and executives approve budgets. Move from execution to strategy to increase your value.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "SEO basics free course (YouTube)", URL: "https://www.youtube.com/results?search_query=seo+for+beginners"},
				{Title: "Google Analytics 4 free course", URL: "https://skillshop.withgoogle.com"},
				{Title: "Digital marketing strategy guide", URL: "https://www.youtube.com/results?search_query=digital+marketing+strategy+nepal"},
			}},
			{StepNumber: 6, Title: "Become a marketing manager or agency owner", Description: "With 3-5 years of experience, you can become a Marketing Manager or Head of Marketing at a company. Or start your own digital marketing agency. Many Nepali agencies now serve international clients from Nepal. Build a team of specialists — content writers, graphic designers, video editors. The digital marketing industry in Nepal is still young and growing fast. You can be part of shaping it. Every business needs marketing and the best marketers are always in demand.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a marketing agency in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=start+marketing+agency+nepal"},
				{Title: "Marketing leadership career path", URL: "https://www.merojob.com"},
				{Title: "Serving international clients from Nepal", URL: "https://www.upwork.com"},
			}},
		},
	}
}

func mobileAppDeveloper() careerSeed {
	return careerSeed{
		CategoryName: "Technology & IT",
		CategorySlug: "technology-it",
		CategoryIcon: "📲",
		Title:        "Mobile App Developer",
		Slug:         "mobile-app-developer",
		Summary:      "Mobile app developers create applications for smartphones and tablets. They build apps for Android and iOS that people use every day.",
		Description:  "A mobile app developer builds applications that run on mobile phones and tablets. They create apps for Android (using Kotlin or Java) and iOS (using Swift). Some developers use cross-platform tools like Flutter or React Native to build apps for both platforms at once. In Nepal, the mobile app development industry is growing as more businesses want their own apps — from e-commerce and banking to food delivery and education. The number of smartphone users in Nepal is increasing rapidly. App developers can work for companies, as freelancers, or build their own apps and earn from ads or sales.",
		DailyTasks: []string{
			"Write code for new features in the mobile app",
			"Fix bugs and errors reported by users",
			"Test the app on different devices and screen sizes",
			"Design user interfaces that are easy to use",
			"Connect the app to backend servers and databases",
			"Review code written by other developers",
			"Publish app updates to Google Play and App Store",
		},
		Skills: []string{
			"Programming in Kotlin/Java (Android) or Swift (iOS)",
			"Cross-platform frameworks (Flutter, React Native)",
			"UI/UX design principles for mobile",
			"Understanding of REST APIs and data formats (JSON)",
			"Version control with Git",
			"App store deployment process",
			"Problem-solving and debugging skills",
		},
		SalaryMin:   300000,
		SalaryMax:   1800000,
		Difficulty:  4,
		FutureProof: 82,
		EducationReq: "Bachelor's in Computer Science or related field preferred but not required. Many successful mobile developers are self-taught or learned through bootcamps. Online courses from Udemy, Coursera, and free resources are widely used.",
		Outlook:      "Mobile app development is growing as smartphone usage increases in Nepal. Companies are investing in mobile apps. The market for Nepali apps in areas like e-commerce, finance, food delivery, and education is expanding rapidly. Remote work opportunities with international companies are growing. AI tools are changing the development process but skilled developers remain essential.",
		Tags:         []string{"tech", "coding", "mobile", "creative", "remote-work"},
		Resources: []resourceSeed{
			{Title: "Flutter Development (Google)", URL: "https://flutter.dev", Description: "Free cross-platform mobile development framework by Google"},
			{Title: "Android Developer Fundamentals", URL: "https://developer.android.com/courses", Description: "Free Android development courses from Google"},
			{Title: "Merojob IT Jobs", URL: "https://www.merojob.com", Description: "Find mobile app developer jobs in Nepal"},
			{Title: "LeetCode - Practice coding", URL: "https://leetcode.com", Description: "Practice coding problems for technical interviews"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn programming basics with a beginner-friendly language", Description: "Start with Python or JavaScript to learn programming concepts: variables, loops, functions, and objects. Then move to Kotlin (for Android) or Swift (for iOS). Or start with Flutter (uses Dart language) which works on both platforms. Build a simple app: a calculator, a to-do list, or a quote generator. Do not worry about making it beautiful. Just make it work. Your first app will be ugly but it is YOUR app. That is something to be proud of.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "FreeCodeCamp - Learn to code free", URL: "https://www.freecodecamp.org"},
				{Title: "Flutter beginners course (YouTube)", URL: "https://www.youtube.com/results?search_query=flutter+for+beginners+2024"},
				{Title: "Kotlin for Android (Google)", URL: "https://developer.android.com/courses"},
			}},
			{StepNumber: 2, Title: "Build 2-3 simple but complete apps", Description: "Build apps that work from start to finish. A weather app that shows real weather data. A note-taking app that saves notes. A simple game. Each app teaches you something new. Put them on GitHub so employers can see your code. Upload them to Google Play (only $25 one-time fee). Having real apps in the store shows you can finish what you start. That matters more than knowing every programming concept.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Beginner app project ideas", URL: "https://www.youtube.com/results?search_query=android+app+ideas+for+beginners"},
				{Title: "How to publish on Google Play", URL: "https://developer.android.com/studio/publish"},
				{Title: "GitHub for mobile developers (YouTube)", URL: "https://www.youtube.com/results?search_query=github+for+mobile+developers"},
			}},
			{StepNumber: 3, Title: "Learn to connect apps to the internet", Description: "Most apps need to talk to a server. Learn REST APIs — how apps send and receive data from the internet. Build an app that shows data from a public API (like news, weather, or movies). Learn to use databases (Firebase is great for beginners). Learn to handle user authentication (login/signup). Apps that talk to servers are real-world apps that companies pay for.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "REST API tutorial for mobile devs (YouTube)", URL: "https://www.youtube.com/results?search_query=rest+api+for+mobile+developers"},
				{Title: "Firebase for Flutter/Android (free)", URL: "https://firebase.google.com"},
				{Title: "How REST APIs work (simple guide)", URL: "https://www.freecodecamp.org/news/rest-api-tutorial/"},
			}},
			{StepNumber: 4, Title: "Build a portfolio of 4-5 quality apps", Description: "Create a portfolio of apps that show different skills: a social media feed, an e-commerce catalog, a map-based app, a chat app. Polish them — good icons, nice design, smooth navigation. Write clear README files for each. Record short demo videos. Your portfolio is what gets you hired. Companies want to see that you can build real, working apps. Quality matters more than quantity.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Mobile app portfolio examples (YouTube)", URL: "https://www.youtube.com/results?search_query=mobile+developer+portfolio"},
				{Title: "UI/UX design principles for mobile", URL: "https://material.io/design"},
				{Title: "App demo video creation guide", URL: "https://www.youtube.com/results?search_query=create+app+demo+video"},
			}},
			{StepNumber: 5, Title: "Apply for jobs or start freelancing", Description: "Update your resume and LinkedIn. Apply to mobile developer positions in Nepal on Merojob and LinkedIn. The best place to find remote work is Upwork, Toptal, or AngelList. Junior developers in Nepal earn 30,000-60,000 NPR per month. With 2-3 years of experience, this goes up to 80,000-150,000 NPR. International remote jobs pay significantly more. Be patient — your first job search may take 2-3 months. Keep building apps while you look.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Find IT jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Remote mobile developer jobs (Upwork)", URL: "https://www.upwork.com"},
				{Title: "Mobile developer interview prep (YouTube)", URL: "https://www.youtube.com/results?search_query=mobile+developer+interview+questions"},
			}},
			{StepNumber: 6, Title: "Go deep — specialize and lead", Description: "After 3-5 years, pick a specialty: games (Unity/Unreal), AR/VR, blockchain, AI/machine learning on mobile, or enterprise app development. Or become a lead developer managing a team. Learn about app architecture, performance optimization, and security. The best mobile developers are lifelong learners — the technologies change every 2-3 years. The opportunity in Nepal is huge because more people get smartphones every day and they all need apps.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Advanced Android development (Google)", URL: "https://developer.android.com/courses"},
				{Title: "Flutter advanced topics", URL: "https://flutter.dev/docs"},
				{Title: "Lead developer career path (YouTube)", URL: "https://www.youtube.com/results?search_query=lead+mobile+developer+role"},
			}},
		},
	}
}

func freelanceWebDeveloper() careerSeed {
	return careerSeed{
		CategoryName: "Technology & IT",
		CategorySlug: "technology-it",
		CategoryIcon: "🌐",
		Title:        "Freelance Web Developer",
		Slug:         "freelance-web-developer",
		Summary:      "Freelance web developers build websites and web applications for clients. They work from home or anywhere, choosing their own projects and hours.",
		Description:  "A freelance web developer builds websites and web applications for clients on a project-by-project basis. Unlike a regular employee, freelancers are self-employed and find their own clients. In Nepal, freelance web development is a growing field because the internet makes it possible to work for clients anywhere in the world. Freelancers build business websites, e-commerce stores, blogs, and custom web applications. Common technologies include HTML, CSS, JavaScript, React, Node.js, PHP, and WordPress. Freelancing offers flexibility and the potential for high earnings, but requires self-discipline, client management skills, and the ability to handle irregular income.",
		DailyTasks: []string{
			"Build and style web pages using HTML, CSS, and JavaScript",
			"Develop website features using frameworks like React or Laravel",
			"Fix bugs and update existing websites for clients",
			"Communicate with clients about project requirements",
			"Set up hosting, domains, and deploy websites live",
			"Send invoices and manage payments",
			"Learn new technologies to stay competitive",
		},
		Skills: []string{
			"HTML, CSS, and JavaScript fundamentals",
			"Frontend frameworks (React, Vue, or Angular)",
			"Backend development (Node.js, PHP, Python, or Ruby)",
			"Database design (MySQL, PostgreSQL, or MongoDB)",
			"WordPress development and customization",
			"Version control with Git",
			"Client communication and business skills",
		},
		SalaryMin:   300000,
		SalaryMax:   2000000,
		Difficulty:  3,
		FutureProof: 78,
		EducationReq: "No degree required. Skills and portfolio are everything. Many successful freelance web developers are self-taught. Online courses from freeCodeCamp, The Odin Project, and Udemy provide all the training needed.",
		Outlook:      "Freelance web development is growing worldwide. Nepal has a growing community of freelancers earning in foreign currency. Platforms like Upwork, Fiverr, and Toptal connect Nepali developers with international clients. The market is competitive but skilled developers can earn very well, especially those who specialize and build a good reputation.",
		Tags:         []string{"tech", "freelance", "coding", "remote-work", "entrepreneur"},
		Resources: []resourceSeed{
			{Title: "freeCodeCamp", URL: "https://www.freecodecamp.org", Description: "Free web development curriculum — start from zero"},
			{Title: "The Odin Project", URL: "https://www.theodinproject.com", Description: "Free full-stack web development course"},
			{Title: "Merojob IT Jobs", URL: "https://www.merojob.com", Description: "Find freelance web development jobs in Nepal"},
			{Title: "Upwork - Freelance platform", URL: "https://www.upwork.com", Description: "Find freelance clients worldwide"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn the web fundamentals — HTML, CSS, JavaScript", Description: "HTML is the structure of web pages. CSS makes them look good. JavaScript makes them interactive. Learn these three first. Build a simple personal website — a page about you, your interests, your goals. Do not use any frameworks yet. Just raw HTML, CSS, and JS. Understanding the fundamentals before using frameworks is like learning to cook before using a recipe app. It makes you a better developer.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "freeCodeCamp - Responsive Web Design", URL: "https://www.freecodecamp.org"},
				{Title: "JavaScript basics (YouTube)", URL: "https://www.youtube.com/results?search_query=javascript+for+beginners+2024"},
				{Title: "The Odin Project - Web Development 101", URL: "https://www.theodinproject.com"},
			}},
			{StepNumber: 2, Title: "Learn a framework and build real projects", Description: "Pick a frontend framework: React (most popular), Vue (easier), or Angular (more structured). Build projects that work: a to-do app, a weather dashboard, a movie search app, a blog. Then learn backend: Node.js/Express or PHP/Laravel. Build a full-stack app where users can sign up, log in, and save data. This is when you become a real web developer. You can build anything now.", Duration: "4-8 months", Links: []roadmapLink{
				{Title: "React tutorial (React.dev)", URL: "https://react.dev/learn"},
				{Title: "Node.js and Express tutorial (YouTube)", URL: "https://www.youtube.com/results?search_query=node+js+for+beginners"},
				{Title: "Full stack open course (University of Helsinki)", URL: "https://fullstackopen.com/en/"},
			}},
			{StepNumber: 3, Title: "Build a portfolio of live websites", Description: "Create 4-5 real projects and put them live on the internet. A portfolio website for yourself. A business website for a friend or local shop (free or low cost). A small e-commerce site. A blog platform. Deploy them on free hosting (Netlify, Vercel, or Render). A live website that people can visit is worth more than a hundred certificates. Show potential clients that you can build real, working websites that load fast and look good on phones.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Netlify - Free web hosting", URL: "https://www.netlify.com"},
				{Title: "Web developer portfolio examples (YouTube)", URL: "https://www.youtube.com/results?search_query=web+developer+portfolio+ideas"},
				{Title: "GitHub Pages - Free portfolio hosting", URL: "https://pages.github.com"},
			}},
			{StepNumber: 4, Title: "Start finding clients on freelance platforms", Description: "Create profiles on Upwork, Fiverr, Freelancer, and PeoplePerHour. In Nepal, also try Merojob and local Facebook groups. Start with small projects — a simple landing page, a WordPress fix, a HTML email. Bid competitively. Your first few jobs may pay very little ($10-50). Do excellent work anyway. Get good reviews. Raise your prices with each project. Five-star reviews are gold on freelance platforms. One happy client leads to many more.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Upwork for Nepali freelancers (YouTube)", URL: "https://www.youtube.com/results?search_query=upwork+nepal+web+developer"},
				{Title: "Fiverr - Start selling web services", URL: "https://www.fiverr.com"},
				{Title: "Freelancing tips for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=freelancing+tips+nepal"},
			}},
			{StepNumber: 5, Title: "Build a personal brand and repeat clients", Description: "Create a professional website for your freelance business. Write blog posts about web development. Share your work on LinkedIn and Twitter. Build relationships with past clients — send them holiday greetings, offer maintenance packages. Repeat clients are the backbone of a freelance business. They trust you and pay better than new clients. Learn to say no to bad projects. Focus on clients who value quality work and pay fairly.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Personal branding for freelancers (YouTube)", URL: "https://www.youtube.com/results?search_query=personal+branding+freelancer+nepal"},
				{Title: "Client relationship tips for freelancers", URL: "https://www.youtube.com/results?search_query=freelance+client+management"},
				{Title: "LinkedIn profile optimization for developers", URL: "https://www.youtube.com/results?search_query=linkedin+for+web+developers"},
			}},
			{StepNumber: 6, Title: "Scale up — agency, products, or niche specialization", Description: "After 2-3 years of freelancing, pick a direction: start a small web development agency hiring other freelancers, build your own SaaS product (software as a service) that generates passive income, or become a highly paid specialist in a niche (Shopify expert, WordPress security, custom API development). Nepal has thousands of freelancers but the ones who specialize and provide exceptional value are the ones who succeed long term.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a web agency in Nepal", URL: "https://www.youtube.com/results?search_query=start+web+agency+nepal"},
				{Title: "SaaS product development guide", URL: "https://www.youtube.com/results?search_query=build+saas+product+as+freelancer"},
				{Title: "Niche specialization for web developers", URL: "https://www.youtube.com/results?search_query=web+developer+niche+specialization"},
			}},
		},
	}
}

func ecommerceEntrepreneur() careerSeed {
	return careerSeed{
		CategoryName: "Technology & IT",
		CategorySlug: "technology-it",
		CategoryIcon: "🛒",
		Title:        "E-commerce Entrepreneur",
		Slug:         "ecommerce-entrepreneur",
		Summary:      "E-commerce entrepreneurs run online stores that sell products to customers over the internet. They manage inventory, marketing, and delivery.",
		Description:  "An e-commerce entrepreneur runs an online store. They sell products — clothes, electronics, handicrafts, food, or anything else — through websites and social media. In Nepal, e-commerce is growing rapidly. More people are shopping online every year. Entrepreneurs can start with a simple Facebook page or Instagram shop, then grow to a full website. Platforms like Daraz, Sastodeal, and Facebook Marketplace make it easy to start selling online. E-commerce entrepreneurs need to handle product sourcing, pricing, marketing, payments, shipping, and customer service. It is a challenging but exciting business that can be started with very little money.",
		DailyTasks: []string{
			"Source products from suppliers or make them yourself",
			"List products on your website and social media",
			"Take orders and process payments",
			"Pack products and arrange delivery or shipping",
			"Respond to customer questions and complaints",
			"Run social media ads to promote your products",
			"Track inventory and reorder popular items",
		},
		Skills: []string{
			"Product sourcing and supplier management",
			"Basic website management (Shopify, WooCommerce)",
			"Social media marketing and online advertising",
			"Customer service and communication",
			"Photography (taking good product photos)",
			"Basic accounting and financial management",
			"Logistics and shipping management",
		},
		SalaryMin:   200000,
		SalaryMax:   2000000,
		Difficulty:  3,
		FutureProof: 75,
		EducationReq: "SLC/SEE pass minimum. Business or marketing training helpful but not required. Many successful e-commerce entrepreneurs learned by doing. Online courses from Shopify, Google, and Facebook are valuable.",
		Outlook:      "E-commerce in Nepal is growing 20-30% per year. More people are shopping online, especially in Kathmandu and other cities. Internet and smartphone penetration are increasing. Cash on delivery is still dominant but digital payments are growing. Competition is increasing but the market is also expanding. Niche products do especially well.",
		Tags:         []string{"tech", "ecommerce", "entrepreneur", "online-business", "marketing"},
		Resources: []resourceSeed{
			{Title: "Shopify - Start an Online Store", URL: "https://www.shopify.com", Description: "Build your e-commerce website easily"},
			{Title: "Sastodeal - Nepal's Online Marketplace", URL: "https://www.sastodeal.com", Description: "Sell products on Nepal's popular online marketplace"},
			{Title: "Daraz Seller Center Nepal", URL: "https://seller.daraz.com.np", Description: "List and sell products on Daraz Nepal"},
			{Title: "Merojob Business Jobs", URL: "https://www.merojob.com", Description: "Find e-commerce jobs and opportunities in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Decide what to sell and find suppliers", Description: "Choose a product category you know and care about: clothes, electronics, home goods, handicrafts, food items, or beauty products. Find suppliers in Nepal — visit wholesale markets in Kathmandu (like New Road, Asan, or Kalimati). For handmade products, work directly with artisans. For imported items, find distributors. Calculate your costs: product price, shipping, packaging, marketing. Make sure you can sell at a profit. Start with 10-20 products, not hundreds.", Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Wholesale markets in Kathmandu guide", URL: "https://www.youtube.com/results?search_query=wholesale+market+kathmandu"},
				{Title: "Product sourcing tips for e-commerce", URL: "https://www.youtube.com/results?search_query=product+sourcing+nepal"},
				{Title: "Sastodeal - Seller registration", URL: "https://www.sastodeal.com"},
			}},
			{StepNumber: 2, Title: "Set up your online store — start simple", Description: "Start with a Facebook Page or Instagram Shop — free and easy. Post good photos of your products with clear prices. Then create a simple website using Shopify, WooCommerce, or a free platform. Take clear, well-lit photos of your products. Write simple descriptions in Nepali and English. Set up payment options: bank transfer, eSewa, Khalti, or cash on delivery. Most Nepali customers prefer cash on delivery for their first purchase. Keep it simple at first. You can improve later.", Duration: "1-2 months", Links: []roadmapLink{
				{Title: "Start selling on Facebook in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=facebook+shop+nepal"},
				{Title: "Shopify setup guide for beginners", URL: "https://www.shopify.com"},
				{Title: "eSewa merchant setup", URL: "https://www.esewa.com.np"},
			}},
			{StepNumber: 3, Title: "Start selling and getting your first customers", Description: "Tell friends and family about your store. Post in Facebook groups related to your products. Offer a small discount for first-time buyers. Ask happy customers to share your page. Respond to messages quickly and politely. Every happy customer is free advertising. Your first 10 customers are the hardest and most important. Treat them like royalty. Ask for feedback and improve based on what they say.",
			Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Facebook marketing for e-commerce (YouTube)", URL: "https://www.youtube.com/results?search_query=facebook+marketing+ecommerce+nepal"},
				{Title: "Customer service tips for online stores", URL: "https://www.youtube.com/results?search_query=online+store+customer+service"},
				{Title: "Instagram for business in Nepal", URL: "https://www.youtube.com/results?search_query=instagram+business+nepal"},
			}},
			{StepNumber: 4, Title: "Scale up with paid advertising", Description: "Once you have some sales and product photos, start running Facebook and Instagram ads. Start with a small budget (500-1000 NPR per day). Target people in Nepal who are interested in your type of products. Track which ads work and which do not. Scale up the ones that work. Learn about Google Shopping ads. Good advertising can bring many customers but you need to make sure your profit per sale is higher than your ad cost.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Facebook Ads for e-commerce (YouTube)", URL: "https://www.youtube.com/results?search_query=facebook+ads+nepal+ecommerce"},
				{Title: "Google Shopping ads setup guide", URL: "https://www.youtube.com/results?search_query=google+shopping+ads+tutorial"},
				{Title: "E-commerce ad budget management", URL: "https://www.youtube.com/results?search_query=ecommerce+ad+budget+nepal"},
			}},
			{StepNumber: 5, Title: "Improve operations — faster delivery, better service", Description: "Set up reliable delivery partnerships. In Kathmandu, you can deliver yourself or use delivery services like Pathao, Foodmandu, or private couriers. For outside Kathmandu, use Nepal Post or private courier companies. Create a system for packing orders quickly and correctly. Build an email list of customers and send them offers. Automate where you can — order confirmations, tracking updates, thank-you messages. Good operations mean happy customers and fewer problems.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Delivery and logistics for e-commerce Nepal", URL: "https://www.youtube.com/results?search_query=delivery+service+nepal+ecommerce"},
				{Title: "Pathao delivery for business", URL: "https://pathao.com/np"},
				{Title: "Email marketing for e-commerce (YouTube)", URL: "https://www.youtube.com/results?search_query=email+marketing+ecommerce"},
			}},
			{StepNumber: 6, Title: "Expand — new products, new channels, new markets", Description: "Add more products that your customers ask for. Sell on multiple platforms: Daraz, Sastodeal, Etsy (for handicrafts internationally), Amazon. Consider exporting Nepali products internationally. Build a brand that people recognize and trust. Hire help when you get too busy. The best e-commerce businesses in Nepal started small and grew step by step. Every big online store was once a single person with a Facebook page and a dream.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Selling on Daraz Nepal seller guide", URL: "https://seller.daraz.com.np"},
				{Title: "Exporting Nepali products online (YouTube)", URL: "https://www.youtube.com/results?search_query=export+nepali+products+online"},
				{Title: "Building an e-commerce brand in Nepal", URL: "https://www.youtube.com/results?search_query=ecommerce+brand+building+nepal"},
			}},
		},
	}
}

func seoSpecialist() careerSeed {
	return careerSeed{
		CategoryName: "Technology & IT",
		CategorySlug: "technology-it",
		CategoryIcon: "🔍",
		Title:        "SEO Specialist",
		Slug:         "seo-specialist",
		Summary:      "SEO specialists help websites rank higher on Google search results. They optimize content, improve technical aspects, and build links to bring more visitors.",
		Description:  "An SEO (Search Engine Optimization) specialist helps websites appear higher in Google search results. When someone searches for something, SEO makes sure your website shows up on the first page. SEO involves researching keywords people search for, creating content around those keywords, improving website speed and structure, and getting other websites to link to yours. In Nepal, SEO is a growing field as more businesses want to be found online. SEO specialists can work for digital marketing agencies, as freelancers, or as in-house experts for companies. It is a technical and analytical job that also requires creativity.",
		DailyTasks: []string{
			"Research keywords that people search for in your industry",
			"Optimize website content and meta tags for keywords",
			"Analyze website traffic using Google Analytics and Search Console",
			"Check website speed and technical issues",
			"Build backlinks through outreach and content creation",
			"Monitor search rankings for target keywords",
			"Write SEO-friendly blog posts and articles",
		},
		Skills: []string{
			"Keyword research tools (Google Keyword Planner, Ahrefs, SEMrush)",
			"On-page SEO (meta tags, headers, content optimization)",
			"Technical SEO (site speed, mobile-friendliness, structured data)",
			"Link building and outreach strategies",
			"Google Analytics and Search Console",
			"Content writing and basic HTML",
			"Analytical thinking and data interpretation",
		},
		SalaryMin:   250000,
		SalaryMax:   1200000,
		Difficulty:  3,
		FutureProof: 75,
		EducationReq: "SLC/SEE pass minimum. SEO certifications from Google, HubSpot, and SEMrush are valuable. Practical experience and case studies matter more than degrees. Digital marketing training institutes in Nepal offer SEO courses.",
		Outlook:      "SEO is a growing field as more businesses compete for online visibility. Google is constantly changing its algorithms, so SEO specialists must keep learning. The rise of AI and voice search is changing how SEO works. Freelance and remote SEO work is available with international clients. Nepali SEO specialists are competitive in the global market due to lower rates and good English skills.",
		Tags:         []string{"tech", "marketing", "analytics", "search-engines", "remote-work"},
		Resources: []resourceSeed{
			{Title: "Google SEO Starter Guide", URL: "https://developers.google.com/search/docs/fundamentals/seo-starter-guide", Description: "Free official Google guide to SEO fundamentals"},
			{Title: "HubSpot SEO Certification", URL: "https://academy.hubspot.com", Description: "Free SEO certification course from HubSpot"},
			{Title: "Merojob Marketing Jobs", URL: "https://www.merojob.com", Description: "Find SEO specialist jobs in Nepal"},
			{Title: "SEMrush SEO Toolkit", URL: "https://www.semrush.com", Description: "Professional SEO tools and learning resources"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn how search engines work", Description: "Understand how Google crawls, indexes, and ranks websites. Learn about search algorithms, ranking factors, and how search results are generated. Read the free Google SEO Starter Guide. Search for things yourself and notice which pages rank first and why. Understanding the basics of how search works is the foundation of everything else in SEO. This knowledge is all available for free on the internet.", Duration: "1-2 months", Links: []roadmapLink{
				{Title: "Google SEO Starter Guide", URL: "https://developers.google.com/search/docs/fundamentals/seo-starter-guide"},
				{Title: "How Google Search works", URL: "https://www.google.com/search/howsearchworks/"},
				{Title: "SEO fundamentals course (YouTube)", URL: "https://www.youtube.com/results?search_query=seo+for+beginners+2024"},
			}},
			{StepNumber: 2, Title: "Learn keyword research and on-page SEO", Description: "Learn how to find keywords that people search for. Use free tools like Google Keyword Planner and AnswerThePublic. Learn to optimize page titles, meta descriptions, headers, and content for target keywords. Study how to write SEO-friendly content that both people and Google love. Practice on your own blog or website. These are the core skills of SEO. Master them before moving to advanced topics.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Google Keyword Planner tutorial", URL: "https://www.youtube.com/results?search_query=google+keyword+planner+tutorial"},
				{Title: "On-page SEO guide for beginners", URL: "https://www.youtube.com/results?search_query=on+page+seo+guide"},
				{Title: "HubSpot SEO Certification (free)", URL: "https://academy.hubspot.com"},
			}},
			{StepNumber: 3, Title: "Get certified in SEO and analytics", Description: "Get free certifications from Google (Google Analytics, Google Search Console), HubSpot (SEO Certification), and SEMrush (SEO Toolkit Certification). These certifications are recognized by employers and clients. Learn Google Analytics 4 (GA4) — it is the industry standard. Learn Google Search Console to monitor website performance in search. Certified SEO specialists are trusted more and can charge higher rates. These certifications cost nothing but time and effort.", Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Google Analytics 4 certification", URL: "https://skillshop.withgoogle.com"},
				{Title: "Google Search Console training", URL: "https://developers.google.com/search/docs"},
				{Title: "SEMrush SEO certification", URL: "https://www.semrush.com/academy/"},
			}},
			{StepNumber: 4, Title: "Build case studies with real projects", Description: "Apply SEO to a real website — your own blog, a friend's business site, or a non-profit. Track improvements in rankings and traffic. Document everything: what you did, what changed, what the results were. A case study showing 'Increased organic traffic by 150% in 3 months' is worth more than any certificate. Real results prove you know what you are doing. Create a portfolio page showing your case studies. This is what gets you hired.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "SEO case study examples (YouTube)", URL: "https://www.youtube.com/results?search_query=seo+case+study+example"},
				{Title: "SEO portfolio tips for beginners", URL: "https://www.youtube.com/results?search_query=seo+portfolio+tips"},
				{Title: "Using Google Search Console for case studies", URL: "https://www.youtube.com/results?search_query=google+search+console+case+study"},
			}},
			{StepNumber: 5, Title: "Learn technical SEO and link building", Description: "Technical SEO covers website speed optimization, mobile-friendliness, structured data (schema markup), sitemaps, robots.txt, and fixing crawl errors. Link building is getting other reputable websites to link to yours — it is one of the most important ranking factors. Learn ethical (white hat) link building: guest posting, broken link building, creating shareable content. Advanced SEO skills command higher pay and are harder to find. This is where you become a specialist rather than a generalist.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Technical SEO guide (YouTube)", URL: "https://www.youtube.com/results?search_query=technical+seo+for+beginners"},
				{Title: "Link building strategies (ethical)", URL: "https://www.youtube.com/results?search_query=white+hat+link+building"},
				{Title: "Schema markup tutorial", URL: "https://www.youtube.com/results?search_query=schema+markup+seo"},
			}},
			{StepNumber: 6, Title: "Specialize or start your own SEO agency", Description: "After 2-3 years, you can become an SEO manager, head of SEO, or start your own SEO agency. Specialize in a niche: local SEO (for Nepali businesses), e-commerce SEO, international SEO (multi-language sites), or SEO for specific industries. Many Nepali SEO freelancers serve international clients and earn in dollars. The SEO field is always changing so you must keep learning. But the fundamentals of understanding what people search for and giving them the best answer — that never changes.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting an SEO agency in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=start+seo+agency+nepal"},
				{Title: "Local SEO for Nepal businesses", URL: "https://www.youtube.com/results?search_query=local+seo+nepal"},
				{Title: "Freelance SEO on Upwork (YouTube)", URL: "https://www.youtube.com/results?search_query=freelance+seo+nepal"},
			}},
		},
	}
}

func cybersecurityAnalyst() careerSeed {
	return careerSeed{
		CategoryName: "Technology & IT",
		CategorySlug: "technology-it",
		CategoryIcon: "🔒",
		Title:        "Cybersecurity Analyst",
		Slug:         "cybersecurity-analyst",
		Summary:      "Cybersecurity analysts protect computers, networks, and data from hackers and cyber attacks. They monitor systems and respond to security threats.",
		Description:  "A cybersecurity analyst protects an organization's computer systems and data from cyber attacks. They monitor networks for suspicious activity, investigate security incidents, install security software (firewalls, antivirus), and train employees about security best practices. In Nepal, cybersecurity is an emerging field. Banks, government offices, e-commerce companies, and telecom companies are most in need of cybersecurity professionals. As more services go online in Nepal, the need for cybersecurity grows. It is a challenging field that requires continuous learning because hackers are always finding new ways to attack.",
		DailyTasks: []string{
			"Monitor network traffic for suspicious activity",
			"Investigate security alerts and potential breaches",
			"Install and configure firewalls and security software",
			"Conduct security audits and vulnerability assessments",
			"Train employees on cybersecurity best practices",
			"Keep security systems and software updated",
			"Document security incidents and responses",
		},
		Skills: []string{
			"Network security fundamentals",
			"Knowledge of operating systems (Windows, Linux)",
			"Understanding of common cyber attacks and defenses",
			"Security tools (firewalls, IDS/IPS, antivirus)",
			"Risk assessment and vulnerability scanning",
			"Analytical thinking and problem-solving",
			"Attention to detail and documentation",
		},
		SalaryMin:   400000,
		SalaryMax:   2000000,
		Difficulty:  4,
		FutureProof: 92,
		EducationReq: "Bachelor's in Computer Science, IT, or Cybersecurity preferred. Certifications like CompTIA Security+, CEH, or CISSP are highly valued. Practical skills and CTF (Capture the Flag) experience matter a lot.",
		Outlook:      "Cybersecurity is one of the fastest growing fields in IT globally. In Nepal, demand is increasing as cyber attacks become more common. Banks, financial institutions, and government agencies are hiring cybersecurity professionals. The field has a severe talent shortage worldwide, which means good salaries and job security. AI is changing the threat landscape but increasing the need for human expertise.",
		Tags:         []string{"tech", "security", "high-growth", "analytics", "protect"},
		Resources: []resourceSeed{
			{Title: "CompTIA Security+ Certification", URL: "https://www.comptia.org/certifications/security", Description: "Industry standard entry-level cybersecurity certification"},
			{Title: "Cybrary - Free Cybersecurity Training", URL: "https://www.cybrary.it", Description: "Free cybersecurity courses and labs"},
			{Title: "Merojob IT Jobs", URL: "https://www.merojob.com", Description: "Find cybersecurity jobs in Nepal"},
			{Title: "TryHackMe - Learn Cyber Security", URL: "https://tryhackme.com", Description: "Free hands-on cybersecurity training platform"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn IT fundamentals and networking", Description: "Before you can defend systems, you need to understand how they work. Learn computer hardware, operating systems (Windows and Linux), and networking (TCP/IP, DNS, HTTP, firewalls). Set up a home lab with virtual machines. Learn to install and configure Windows Server and Linux. Understanding the basics of IT is essential for cybersecurity. You cannot protect what you do not understand. Free resources like Professor Messer and Cybrary are great places to start.", Duration: "4-6 months", Links: []roadmapLink{
				{Title: "Professor Messer - Free CompTIA A+ training", URL: "https://www.professormesser.com"},
				{Title: "Linux basics for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=linux+for+beginners"},
				{Title: "Networking fundamentals course", URL: "https://www.youtube.com/results?search_query=networking+fundamentals+for+cybersecurity"},
			}},
			{StepNumber: 2, Title: "Learn cybersecurity fundamentals and get certified", Description: "Learn about common threats: malware, phishing, ransomware, DDoS attacks, social engineering. Understand the CIA triad (Confidentiality, Integrity, Availability). Get the CompTIA Security+ certification — it is the entry-level standard for cybersecurity. Study for 2-3 months and take the exam. Security+ teaches you the core concepts and is recognized by employers worldwide. It is the first step on your cybersecurity career path.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "CompTIA Security+ study guide", URL: "https://www.comptia.org/certifications/security"},
				{Title: "Professor Messer - Free Security+ training", URL: "https://www.professormesser.com"},
				{Title: "Cybrary - Free cybersecurity courses", URL: "https://www.cybrary.it"},
			}},
			{StepNumber: 3, Title: "Practice with hands-on labs and CTF challenges", Description: "Theory alone is not enough. Set up a home lab with virtual machines. Practice on TryHackMe and HackTheBox — they have free cybersecurity challenges. Participate in Capture The Flag (CTF) competitions. Learn to use tools like Nmap (network scanning), Wireshark (packet analysis), Metasploit (penetration testing), and Burp Suite (web security testing). Hands-on practice is what separates real cybersecurity professionals from people who just have certifications.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "TryHackMe - Free cyber security training", URL: "https://tryhackme.com"},
				{Title: "HackTheBox - Practice platform", URL: "https://www.hackthebox.com"},
				{Title: "Wireshark tutorial for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=wireshark+tutorial+for+beginners"},
			}},
			{StepNumber: 4, Title: "Get your first cybersecurity job or internship", Description: "Apply for entry-level cybersecurity roles: Security Analyst, SOC (Security Operations Center) Analyst, or Junior Penetration Tester. In Nepal, you may start in an IT support or network admin role and move into security. Emphasize your certifications, home lab experience, and CTF participation. Be honest about your skill level but show enthusiasm for learning. The cybersecurity field is desperate for talent — if you have the skills and drive, you will find opportunities. Many security professionals started in other IT roles and transitioned.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Find cybersecurity jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Entry level cybersecurity resume tips (YouTube)", URL: "https://www.youtube.com/results?search_query=entry+level+cybersecurity+resume"},
				{Title: "SOC analyst role explained", URL: "https://www.youtube.com/results?search_query=what+is+a+soc+analyst"},
			}},
			{StepNumber: 5, Title: "Specialize and get advanced certifications", Description: "After 1-2 years of experience, choose a specialization: penetration testing (ethical hacking), security architecture, incident response, governance and compliance, or cloud security. Get advanced certifications: CEH (Certified Ethical Hacker), CISSP (Certified Information Systems Security Professional), or OSCP (Offensive Security Certified Professional). These certifications open doors to senior roles and significantly higher pay. Specialization is the key to career growth in cybersecurity.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "CEH certification guide", URL: "https://www.eccouncil.org/programs/certified-ethical-hacker-ceh/"},
				{Title: "CISSP certification overview", URL: "https://www.isc2.org/Certifications/CISSP"},
				{Title: "OSCP - Offensive Security certification", URL: "https://www.offensive-security.com/pwk-oscp/"},
			}},
			{StepNumber: 6, Title: "Become a security leader or consultant", Description: "With 5+ years of experience, become a Security Manager, Chief Information Security Officer (CISO), or cybersecurity consultant. Many companies in Nepal need cybersecurity expertise but cannot afford full-time staff — they hire consultants. You can consult for multiple organizations, conduct security audits, and help companies build their security posture. Cybersecurity is a field where your skills only become more valuable with time. As technology advances, the need for security grows — and you will be there to protect it.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Cybersecurity management career path", URL: "https://www.youtube.com/results?search_query=cybersecurity+manager+career"},
				{Title: "Starting a cybersecurity consulting business", URL: "https://www.youtube.com/results?search_query=cybersecurity+consultant+business"},
				{Title: "CISO role and responsibilities (YouTube)", URL: "https://www.youtube.com/results?search_query=what+is+a+ciso"},
			}},
		},
	}
}

func droneOperator() careerSeed {
	return careerSeed{
		CategoryName: "Technology & IT",
		CategorySlug: "technology-it",
		CategoryIcon: "🛸",
		Title:        "Drone Operator",
		Slug:         "drone-operator",
		Summary:      "Drone operators fly unmanned aerial vehicles (drones) for photography, agriculture, mapping, surveying, and inspection purposes.",
		Description:  "A drone operator flies drones for various commercial purposes. In Nepal, drones are used for aerial photography and videography (especially for tourism), agricultural crop monitoring, land surveying, construction site inspection, disaster response, and infrastructure inspection (power lines, bridges, roads). The drone industry in Nepal is growing as the technology becomes more affordable and the government creates clearer regulations. Drone operators need a license from the Civil Aviation Authority of Nepal (CAAN). Many drone operators work as freelancers, offering their services to tourism companies, construction firms, and government agencies.",
		DailyTasks: []string{
			"Plan flight paths for drone missions",
			"Conduct pre-flight checks on drone equipment",
			"Fly drones for photography, surveying, or inspection",
			"Process images and videos after flights",
			"Create maps, 3D models, or inspection reports",
			"Maintain and repair drone equipment",
			"Follow safety regulations and keep flight logs",
		},
		Skills: []string{
			"Drone piloting skills and flight safety",
			"Knowledge of drone regulations in Nepal (CAAN)",
			"Aerial photography and videography skills",
			"Photo and video editing (Lightroom, Premiere Pro)",
			"Basic knowledge of mapping and surveying software",
			"Attention to detail and safety awareness",
			"Customer service and communication",
		},
		SalaryMin:   250000,
		SalaryMax:   1200000,
		Difficulty:  3,
		FutureProof: 72,
		EducationReq: "SLC/SEE pass minimum. Must obtain Drone Pilot License from CAAN Nepal. Training from CAAN-approved institutes. Photography/videography skills are valuable. Engineering background helpful for surveying/inspection work.",
		Outlook:      "The drone industry in Nepal is in its early stages but growing fast. Applications in agriculture, tourism, construction, and disaster management are expanding. The government is developing clearer regulations which will help the industry grow. Competition is still low but will increase. First movers have a big advantage.",
		Tags:         []string{"tech", "drones", "aviation", "photography", "outdoor"},
		Resources: []resourceSeed{
			{Title: "Civil Aviation Authority Nepal - Drone Regulations", URL: "https://www.caanepal.gov.np", Description: "Official drone licensing and regulation information"},
			{Title: "DJI - Drone Technology", URL: "https://www.dji.com", Description: "World's leading drone manufacturer with training resources"},
			{Title: "Merojob Technology Jobs", URL: "https://www.merojob.com", Description: "Find drone operator jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn to fly drones safely", Description: "Start with a cheap beginner drone (under 20,000 NPR) and practice flying in open areas. Learn basic flight controls, how to hover, how to navigate, and how to land smoothly. Learn about drone safety: keep line of sight, avoid people and animals, do not fly near airports. Watch free YouTube tutorials. Practice every day until flying feels natural. A good drone operator can fly smoothly without thinking about the controls.",
			Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Drone flying tips for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+fly+a+drone+for+beginners"},
				{Title: "DJI drone flight tutorial", URL: "https://www.dji.com"},
				{Title: "Drone safety guidelines (YouTube)", URL: "https://www.youtube.com/results?search_query=drone+safety+tips"},
			}},
			{StepNumber: 2, Title: "Get your drone pilot license from CAAN Nepal", Description: "The Civil Aviation Authority of Nepal requires all commercial drone operators to have a Remote Pilot License. The process includes training at a CAAN-approved institute, passing a written exam, and a practical flight test. The training covers aviation regulations, meteorology, navigation, flight planning, and safety. Having a proper license makes you legal and builds trust with clients. Flying without a license can result in fines or confiscation of your drone.",
			Duration: "2-4 months", Links: []roadmapLink{
				{Title: "CAAN - Drone pilot licensing info", URL: "https://www.caanepal.gov.np"},
				{Title: "Drone training institutes in Nepal (CAAN approved)", URL: "https://www.caanepal.gov.np"},
				{Title: "How to get drone license Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=drone+license+nepal+caan"},
			}},
			{StepNumber: 3, Title: "Learn aerial photography and video skills", Description: "Most drone jobs involve photography or videography. Learn composition rules for aerial shots. Practice different camera movements: flyover, orbit, reveal, follow. Learn to edit photos in Lightroom and videos in Premiere Pro. Create a demo reel of your best aerial footage. Aerial photography skills separate a basic drone pilot from a professional who gets paid well. Tourist resorts, hotels, and real estate agents are always looking for good aerial photos and videos.",
			Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Aerial photography tips for drones (YouTube)", URL: "https://www.youtube.com/results?search_query=aerial+photography+tips+drone"},
				{Title: "Lightroom editing for drone photos", URL: "https://www.youtube.com/results?search_query=lightroom+drone+photo+editing"},
				{Title: "Premiere Pro video editing for drone footage", URL: "https://www.youtube.com/results?search_query=premiere+pro+drone+editing"},
			}},
			{StepNumber: 4, Title: "Find your first clients and build a portfolio", Description: "Start by offering free or low-cost drone services to local businesses — hotels wanting aerial photos, real estate agents needing property shots, or construction companies wanting progress photos. Build a portfolio of your best work. Create a simple website or Facebook page. Join local business groups and offer your services. Word of mouth is powerful in Nepal. One happy client leads to referrals. Be professional, reliable, and deliver high-quality work on time.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Marketing drone services in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=drone+business+nepal"},
				{Title: "Building a drone portfolio (examples)", URL: "https://www.youtube.com/results?search_query=drone+photography+portfolio"},
				{Title: "Drone service pricing guide", URL: "https://www.youtube.com/results?search_query=how+much+to+charge+drone+services"},
			}},
			{StepNumber: 5, Title: "Expand into specialized drone services", Description: "Learn specialized skills that pay more: agricultural drone spraying, 3D mapping and modeling (using software like Pix4D or DroneDeploy), thermal imaging for solar panel or building inspection, search and rescue, or cinematography for film production. Each specialty requires additional training but offers higher pay and less competition. Drone mapping for construction and mining is especially in demand in Nepal as infrastructure projects grow.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Agricultural drone spraying guide (YouTube)", URL: "https://www.youtube.com/results?search_query=agricultural+drone+spraying+nepal"},
				{Title: "Drone mapping and surveying course", URL: "https://www.youtube.com/results?search_query=drone+mapping+tutorial"},
				{Title: "Thermal drone inspection training", URL: "https://www.youtube.com/results?search_query=thermal+drone+inspection"},
			}},
			{StepNumber: 6, Title: "Grow your drone business or company", Description: "Build a team of drone operators and offer comprehensive services. Buy more drones for different purposes (agricultural spray drones, mapping drones, cinema drones). Work on government and NGO projects for larger contracts. Develop a specialty that makes you the go-to person in Nepal. The drone industry is still young in Nepal and there is a huge opportunity for early adopters who build a reputation for professional, safe, and high-quality work.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Growing a drone business in Nepal", URL: "https://www.youtube.com/results?search_query=drone+business+growth+nepal"},
				{Title: "Nepal government drone projects", URL: "https://www.caanepal.gov.np"},
				{Title: "Drone entrepreneurship guide (YouTube)", URL: "https://www.youtube.com/results?search_query=drone+startup+nepal"},
			}},
		},
	}
}

func pharmacist() careerSeed {
	return careerSeed{
		CategoryName: "Healthcare & Wellness",
		CategorySlug: "healthcare-wellness",
		CategoryDesc: "Careers focused on health, medicine, and helping people feel better physically and mentally.",
		CategoryIcon: "💊",
		Title:        "Pharmacist",
		Slug:         "pharmacist",
		Summary:      "Pharmacists prepare and dispense medicines. They advise patients on how to take medicines safely and help people manage their health.",
		Description:  "A pharmacist is a healthcare professional who dispenses medicines and advises patients on their safe use. In Nepal, pharmacists work in private pharmacies, hospital pharmacies, pharmaceutical companies, and wholesale drug distributors. They must understand how different medicines work, their side effects, and interactions with other drugs. Pharmacists also educate patients about proper dosage and storage of medicines. Nepal has strict regulations for pharmacy practice through the Nepal Pharmacy Council. The pharmaceutical sector in Nepal is growing with many local manufacturing companies. Pharmacists play a vital role in Nepal's healthcare system, especially in rural areas where they are often the most accessible healthcare provider.",
		DailyTasks: []string{
			"Dispense prescribed medicines to patients",
			"Check prescriptions for errors or interactions",
			"Advise patients on proper medicine usage and dosage",
			"Manage pharmacy inventory and order supplies",
			"Keep records of controlled substances",
			"Counsel patients on managing their health conditions",
			"Prepare compounded medications when needed",
		},
		Skills: []string{
			"Knowledge of medicines, their uses and side effects",
			"Understanding of drug interactions and contraindications",
			"Attention to detail and accuracy",
			"Customer service and patient counseling",
			"Inventory management and record keeping",
			"Knowledge of Nepal drug regulations",
			"Communication skills in Nepali and English",
		},
		SalaryMin:   300000,
		SalaryMax:   1000000,
		Difficulty:  3,
		FutureProof: 82,
		EducationReq: "Bachelor of Pharmacy (B.Pharm) degree from a recognized university (Tribhuvan University, Pokhara University, or Kathmandu University). Must register with Nepal Pharmacy Council. Diploma in Pharmacy (D.Pharm) allows working in lower-level roles.",
		Outlook:      "Pharmacists are always in demand because people always need medicines. The pharmaceutical industry in Nepal is growing with more local manufacturing. Hospital pharmacies and private pharmacy chains are expanding. Rural areas especially need pharmacists. The Nepal Pharmacy Council regulates the profession and ensures quality standards.",
		Tags:         []string{"healthcare", "medicine", "helping-people", "stable", "science"},
		Resources: []resourceSeed{
			{Title: "Nepal Pharmacy Council", URL: "https://www.nepalpharmacycouncil.org.np", Description: "Official regulatory body for pharmacy professionals in Nepal"},
			{Title: "Department of Drug Administration Nepal", URL: "https://www.dda.gov.np", Description: "Government agency regulating medicines and pharmacy practice"},
			{Title: "Merojob Healthcare Jobs", URL: "https://www.merojob.com", Description: "Find pharmacist jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Study science in high school with biology, chemistry, and physics", Description: "To become a pharmacist, you need a strong science background. Focus on chemistry (organic and inorganic), biology, and physics in your SEE and +2 level. Aim for good grades — pharmacy programs are competitive. Read about medicines and how they work. Visit a local pharmacy and ask the pharmacist about their job. Understanding the path early helps you stay motivated. Every journey starts with the first step and for you that is science class.", Duration: "2 years", Links: []roadmapLink{
				{Title: "Tribhuvan University - Pharmacy program info", URL: "https://www.tu.edu.np"},
				{Title: "Kathmandu University - Pharmacy courses", URL: "https://www.ku.edu.np"},
				{Title: "Nepal Pharmacy Council - Approved colleges", URL: "https://www.nepalpharmacycouncil.org.np"},
			}},
			{StepNumber: 2, Title: "Complete a Bachelor of Pharmacy degree", Description: "Enroll in a B.Pharm program at a recognized university in Nepal: Tribhuvan University, Pokhara University, Kathmandu University, or Purbanchal University. The program takes 4 years. You will study pharmacology, pharmaceutical chemistry, pharmacognosy (medicinal plants), pharmacy practice, and hospital pharmacy. Do internships in hospital and community pharmacies during your studies. Pay attention in pharmacology — it is the heart of pharmacy. Form study groups with classmates. Help each other succeed.", Duration: "4 years", Links: []roadmapLink{
				{Title: "Pokhara University - B.Pharm program", URL: "https://www.pu.edu.np"},
				{Title: "Purbanchal University - Pharmacy", URL: "https://www.purbuniv.edu"},
				{Title: "Free pharmacology resources (YouTube)", URL: "https://www.youtube.com/results?search_query=pharmacology+for+beginners"},
			}},
			{StepNumber: 3, Title: "Register with the Nepal Pharmacy Council and get licensed", Description: "After graduation, register with the Nepal Pharmacy Council to get your professional license. The council requires submission of your degree, internship completion, and payment of registration fees. You must also pass a licensing exam. Once registered, you can legally practice as a pharmacist in Nepal. Registration must be renewed periodically with continuing education credits. This license is your passport to practice. Keep it current and follow the professional code of conduct.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Nepal Pharmacy Council - Registration process", URL: "https://www.nepalpharmacycouncil.org.np"},
				{Title: "Licensing exam preparation guide", URL: "https://www.youtube.com/results?search_query=pharmacy+license+exam+nepal"},
				{Title: "Continuing pharmacy education in Nepal", URL: "https://www.dda.gov.np"},
			}},
			{StepNumber: 4, Title: "Find your first pharmacy job", Description: "Apply for positions in community pharmacies, hospital pharmacies, or pharmaceutical companies. In Nepal, many pharmacists work in private pharmacies. Hospital pharmacy jobs are more prestigious and offer better pay. Update your resume with your education and any internship experience. Be prepared for interviews where they will test your drug knowledge. Your first job may not pay much but the experience is invaluable. Learn from senior pharmacists. Be reliable and accurate. In pharmacy, accuracy is literally a matter of life and death.", Duration: "1-6 months", Links: []roadmapLink{
				{Title: "Find pharmacy jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Hospital pharmacy vs community pharmacy", URL: "https://www.youtube.com/results?search_query=hospital+pharmacy+career+nepal"},
				{Title: "Pharmacy interview questions and answers", URL: "https://www.youtube.com/results?search_query=pharmacist+interview+questions"},
			}},
			{StepNumber: 5, Title: "Specialize or pursue higher education", Description: "Consider a Master's in Pharmacy (M.Pharm) in a specialization like pharmacology, pharmaceutical chemistry, or hospital pharmacy. Specialization leads to higher pay and better positions. Some pharmacists move into pharmaceutical company roles: quality control, regulatory affairs, or medical marketing. Others open their own pharmacy business. The pharmacy field offers many paths — clinical, industrial, academic, or entrepreneurial. Find what interests you most and go deeper.", Duration: "2-3 years", Links: []roadmapLink{
				{Title: "M.Pharm programs in Nepal", URL: "https://www.tu.edu.np"},
				{Title: "Pharmaceutical industry careers (YouTube)", URL: "https://www.youtube.com/results?search_query=pharmaceutical+industry+career+nepal"},
				{Title: "How to open a pharmacy in Nepal", URL: "https://www.dda.gov.np"},
			}},
			{StepNumber: 6, Title: "Become a senior pharmacist or pharmacy owner", Description: "With 5+ years of experience, become a senior pharmacist, pharmacy manager, or open your own pharmacy chain. Experienced pharmacists are respected members of the community. In Nepal, trusted pharmacists often serve the same families for generations. The pharmacy profession offers stable income, respect, and the satisfaction of helping people every day. Medicines heal people and as a pharmacist you are a key part of that healing process.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Pharmacy management tips (YouTube)", URL: "https://www.youtube.com/results?search_query=pharmacy+management+nepal"},
				{Title: "Nepal Pharmaceutical Association", URL: "https://www.nepalpharmacycouncil.org.np"},
				{Title: "Opening a pharmacy chain in Nepal", URL: "https://www.dda.gov.np"},
			}},
		},
	}
}

func labTechnician() careerSeed {
	return careerSeed{
		CategoryName: "Healthcare & Wellness",
		CategorySlug: "healthcare-wellness",
		CategoryIcon: "🔬",
		Title:        "Lab Technician",
		Slug:         "lab-technician",
		Summary:      "Lab technicians analyze blood, urine, and other samples to help doctors diagnose diseases and monitor patient health.",
		Description:  "A lab technician (also called a medical laboratory technologist) analyzes patient samples like blood, urine, and tissue to help doctors diagnose diseases. They operate laboratory equipment, prepare samples, perform tests, and record results. In Nepal, lab technicians work in hospital laboratories, private diagnostic centers, public health laboratories, and research facilities. The job requires careful attention to detail and accuracy because test results guide medical decisions. With the growth of healthcare in Nepal, the demand for qualified lab technicians is increasing. There is a particular shortage of lab technicians in rural areas.",
		DailyTasks: []string{
			"Collect blood, urine, and other samples from patients",
			"Prepare samples for testing using proper procedures",
			"Operate lab equipment like microscopes and analyzers",
			"Perform chemical, biological, and microbiological tests",
			"Record and verify test results accurately",
			"Maintain lab equipment and order supplies",
			"Follow safety and infection control protocols",
		},
		Skills: []string{
			"Knowledge of laboratory techniques and procedures",
			"Operation of lab equipment (microscopes, centrifuges, analyzers)",
			"Sample collection and preparation skills",
			"Attention to detail and accuracy",
			"Understanding of infection control and lab safety",
			"Basic computer skills for data entry",
			"Communication with doctors and patients",
		},
		SalaryMin:   250000,
		SalaryMax:   800000,
		Difficulty:  2,
		FutureProof: 78,
		EducationReq: "Diploma in Medical Lab Technology (DMLT) or Bachelor's in Medical Lab Technology (BMLT) from recognized institutes like CTEVT or affiliated universities. Certification from Nepal Health Professional Council required.",
		Outlook:      "The demand for lab technicians in Nepal is growing as the healthcare sector expands. New hospitals and diagnostic centers are opening regularly. Government health posts also need lab technicians. The field offers stable employment. Automation is changing some lab work but skilled technicians are still needed to operate equipment and interpret results.",
		Tags:         []string{"healthcare", "medical", "science", "stable", "helping-people"},
		Resources: []resourceSeed{
			{Title: "CTEVT Nepal - Lab Technology Programs", URL: "https://www.ctevt.org.np", Description: "Council for Technical Education and Vocational Training - DMLT programs"},
			{Title: "Nepal Health Professional Council", URL: "https://www.nhpc.gov.np", Description: "Regulatory body for health professionals in Nepal"},
			{Title: "Merojob Healthcare Jobs", URL: "https://www.merojob.com", Description: "Find lab technician jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Study science subjects in high school", Description: "Focus on biology, chemistry, and physics in your SEE and +2 level. Good grades in science are important for admission into lab technology programs. Read about how laboratories work and what lab technicians do. Visit a diagnostic center and ask if you can observe for a day. Understanding the daily work of a lab technician will help you decide if this career is right for you. The work is detailed and repetitive — you need to enjoy precision and accuracy.", Duration: "1-2 years", Links: []roadmapLink{
				{Title: "CTEVT - Lab Technology course info", URL: "https://www.ctevt.org.np"},
				{Title: "Career as a lab technician (YouTube)", URL: "https://www.youtube.com/results?search_query=medical+lab+technician+career+nepal"},
				{Title: "Biology for lab technicians (free resources)", URL: "https://www.youtube.com/results?search_query=biology+basics+for+lab+tech"},
			}},
			{StepNumber: 2, Title: "Complete a Diploma or Bachelor's in Lab Technology", Description: "Enroll in a DMLT (3 years) or BMLT (4 years) program at a CTEVT-affiliated institute or university. You will study hematology (blood), clinical biochemistry, microbiology, pathology, and immunology. Practical lab work is a major part of the training. Pay close attention to lab safety and quality control procedures. Mistakes in the lab can have serious consequences for patients. Develop a habit of double-checking everything.", Duration: "3-4 years", Links: []roadmapLink{
				{Title: "CTEVT - List of affiliated colleges", URL: "https://www.ctevt.org.np"},
				{Title: "Tribhuvan University - BMLT program", URL: "https://www.tu.edu.np"},
				{Title: "Lab safety training (YouTube)", URL: "https://www.youtube.com/results?search_query=lab+safety+for+beginners"},
			}},
			{StepNumber: 3, Title: "Get licensed and start your internship", Description: "Register with the Nepal Health Professional Council after graduation. Complete a mandatory internship at a hospital laboratory or diagnostic center. Internship gives you real-world experience under supervision. Learn to handle real patient samples, operate advanced equipment, and work under time pressure. This is where you transition from a student to a professional. Make the most of every learning opportunity in your internship.", Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Nepal Health Professional Council - Registration", URL: "https://www.nhpc.gov.np"},
				{Title: "Lab internship tips (YouTube)", URL: "https://www.youtube.com/results?search_query=medical+lab+internship+experience"},
				{Title: "Quality control in medical laboratories", URL: "https://www.youtube.com/results?search_query=lab+quality+control+procedures"},
			}},
			{StepNumber: 4, Title: "Find a job in a hospital or diagnostic center", Description: "Apply for lab technician positions in hospitals, diagnostic centers, and public health labs. Emphasize your training, internship experience, and any specialized skills. Entry-level jobs may be in small clinics or diagnostic centers. Government hospitals offer stable jobs with benefits but require passing the Public Service Commission (Loksewa) exam. Private hospitals and diagnostic chains often pay better. Be willing to work in different departments — hematology, biochemistry, microbiology — to build broad experience.", Duration: "1-6 months", Links: []roadmapLink{
				{Title: "Find lab technician jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Loksewa exam for lab technicians", URL: "https://www.psc.gov.np"},
				{Title: "Hospital lab vs diagnostic center careers", URL: "https://www.youtube.com/results?search_query=lab+technician+work+nepal"},
			}},
			{StepNumber: 5, Title: "Specialize in an advanced lab area", Description: "After 2-3 years, specialize in a higher-demand area: histopathology (tissue analysis), cytology (cell analysis), microbiology, or blood banking (transfusion medicine). Get additional certifications in your chosen specialty. Specialized lab technicians earn more and have better job opportunities. Some technicians move into lab management, quality assurance, or medical equipment sales. The healthcare field offers many paths for growth. Find what interests you and go deeper.", Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Specialization options for lab techs (YouTube)", URL: "https://www.youtube.com/results?search_query=medical+lab+specializations"},
				{Title: "Histopathology training resources", URL: "https://www.youtube.com/results?search_query=histopathology+for+beginners"},
				{Title: "Blood bank technology training", URL: "https://www.youtube.com/results?search_query=blood+bank+technology+training"},
			}},
			{StepNumber: 6, Title: "Advance to supervisor, manager, or teacher", Description: "With 5+ years of experience, become a lab supervisor or manager. Supervise other technicians, manage quality control programs, and oversee lab operations. Some experienced technicians teach at CTEVT institutes or universities. Others open their own diagnostic lab. The healthcare industry in Nepal is growing and the demand for quality diagnostic services is increasing. Your careful work saves lives every day — even if the patients never see your face.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Lab management career path (YouTube)", URL: "https://www.youtube.com/results?search_query=medical+lab+management"},
				{Title: "Starting a diagnostic lab in Nepal", URL: "https://www.dda.gov.np"},
				{Title: "Teaching career in lab technology", URL: "https://www.ctevt.org.np"},
			}},
		},
	}
}

func ayurvedaDoctor() careerSeed {
	return careerSeed{
		CategoryName: "Healthcare & Wellness",
		CategorySlug: "healthcare-wellness",
		CategoryIcon: "🌿",
		Title:        "Ayurveda Doctor (Ayurvedacharya)",
		Slug:         "ayurveda-doctor",
		Summary:      "Ayurveda doctors use traditional Nepali and Indian herbal medicine to treat patients. They prescribe natural remedies, diet changes, and lifestyle advice.",
		Description:  "An Ayurveda doctor (Ayurvedacharya) practices Ayurveda, the traditional system of medicine that originated in South Asia thousands of years ago. Ayurveda uses natural herbs, dietary changes, massages, and lifestyle recommendations to treat illness and promote health. In Nepal, Ayurveda is an officially recognized medical system with government hospitals, research centers, and universities offering degrees. The Ministry of Ayurveda and Alternative Medicine oversees the practice. Many Nepalis prefer Ayurveda for certain conditions, especially chronic diseases, digestive issues, and joint pain. Ayurveda doctors work in government Ayurveda hospitals, private clinics, pharmaceutical companies, and wellness centers. The global interest in natural medicine is creating new opportunities.",
		DailyTasks: []string{
			"Examine patients using Ayurvedic diagnostic methods (pulse, tongue, eyes)",
			"Prescribe herbal medicines and treatments",
			"Advise patients on diet (pathya) and lifestyle changes",
			"Perform Panchakarma therapies (detoxification treatments)",
			"Prepare and dispense herbal formulations",
			"Keep patient records and track treatment progress",
			"Conduct health awareness programs in the community",
		},
		Skills: []string{
			"In-depth knowledge of Ayurvedic principles and treatments",
			"Herbal medicine identification and preparation",
			"Diagnostic skills (Nadi pariksha - pulse diagnosis)",
			"Panchakarma therapy techniques",
			"Patient counseling and communication",
			"Knowledge of medicinal plants found in Nepal",
			"Basic understanding of modern medicine for referrals",
		},
		SalaryMin:   250000,
		SalaryMax:   1200000,
		Difficulty:  4,
		FutureProof: 72,
		EducationReq: "Bachelor of Ayurvedic Medicine and Surgery (BAMS) from a recognized university like Tribhuvan University or Rajiv Gandhi University. Must register with the Nepal Ayurveda Medical Council.",
		Outlook:      "Ayurveda is growing in Nepal and globally as people seek natural healthcare. The Government of Nepal operates Ayurveda hospitals and health centers nationwide. The tourism industry creates demand for Ayurveda in wellness centers and retreats. Nepal's rich biodiversity provides many medicinal plants. Competition from modern medicine exists but Ayurveda has a loyal patient base.",
		Tags:         []string{"healthcare", "ayurveda", "herbal", "traditional", "wellness"},
		Resources: []resourceSeed{
			{Title: "Nepal Ayurveda Medical Council", URL: "https://www.namc.org.np", Description: "Regulatory body for Ayurveda practitioners in Nepal"},
			{Title: "Department of Ayurveda and Alternative Medicine", URL: "https://www.mohp.gov.np", Description: "Government department overseeing Ayurveda services in Nepal"},
			{Title: "Merojob Healthcare Jobs", URL: "https://www.merojob.com", Description: "Find Ayurveda doctor jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn about Ayurveda and its principles", Description: "Read about the basics of Ayurveda: the three doshas (Vata, Pitta, Kapha), the five elements, and how Ayurveda views health and disease. Visit an Ayurveda clinic or hospital. Talk to Ayurveda doctors about their work. Learn about the medicinal plants in your area — many grow wild in Nepal. A strong interest in natural healing is the first requirement for this path. Read free resources online from the Department of Ayurveda.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Department of Ayurveda Nepal - Resources", URL: "https://www.mohp.gov.np"},
				{Title: "Ayurveda basics for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=ayurveda+for+beginners"},
				{Title: "Medicinal plants of Nepal guide", URL: "https://www.youtube.com/results?search_query=medicinal+plants+nepal"},
			}},
			{StepNumber: 2, Title: "Complete your BAMS degree", Description: "Enroll in a Bachelor of Ayurvedic Medicine and Surgery (BAMS) program. In Nepal, Tribhuvan University's Institute of Medicine offers BAMS. The program takes 5.5 years including internship. Study Ayurvedic philosophy, anatomy, physiology, pharmacology (Dravyaguna), toxicology, Panchakarma, and surgery. Also study modern medical subjects to understand both systems. BAMS is an intense program but graduates are respected as fully qualified doctors. The combination of ancient wisdom and modern science makes you a complete healer.", Duration: "5.5 years", Links: []roadmapLink{
				{Title: "Tribhuvan University - BAMS program", URL: "https://www.tu.edu.np"},
				{Title: "Rajiv Gandhi University - Ayurveda programs", URL: "https://www.rgu.edu.np"},
				{Title: "BAMS curriculum and study resources (YouTube)", URL: "https://www.youtube.com/results?search_query=bams+study+tips"},
			}},
			{StepNumber: 3, Title: "Get licensed with the Nepal Ayurveda Medical Council", Description: "After completing BAMS, register with the Nepal Ayurveda Medical Council (NAMC). Submit your degree, internship certificate, and other required documents. Pass the licensing exam. Once registered, you can legally practice Ayurveda in Nepal. The council also requires continuing education for license renewal. This license is essential for your career. With it, you can work in government hospitals, open your own clinic, or work in the private sector.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Nepal Ayurveda Medical Council - Registration", URL: "https://www.namc.org.np"},
				{Title: "Ayurveda licensing exam tips (YouTube)", URL: "https://www.youtube.com/results?search_query=ayurveda+license+exam+nepal"},
				{Title: "Continuing education for Ayurveda doctors", URL: "https://www.mohp.gov.np"},
			}},
			{StepNumber: 4, Title: "Start practicing and gaining experience", Description: "Join a government Ayurveda hospital, private clinic, or wellness center. Government positions offer stability and benefits. Private practice can be more lucrative. Start treating common conditions: digestive issues, arthritis, skin problems, stress, and chronic diseases. Build your reputation through good results and compassionate care. Keep learning about new treatments and herbs. Document interesting cases. Experience is the best teacher in Ayurveda — every patient is different and teaches you something new.", Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Find Ayurveda jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Government Ayurveda hospital careers", URL: "https://www.mohp.gov.np"},
				{Title: "Private Ayurveda practice tips (YouTube)", URL: "https://www.youtube.com/results?search_query=ayurveda+private+practice+nepal"},
			}},
			{StepNumber: 5, Title: "Specialize in Panchakarma or herbal medicine", Description: "Specialize in Panchakarma therapy (the Ayurvedic detoxification system) which is highly demanded by wellness tourists. Or specialize in herbal pharmacology — learn to identify, prepare, and formulate herbal medicines. Get additional training and certification. Specialists command higher fees and have more opportunities. Nepal's biodiversity is a treasure trove of medicinal plants. Many Ayurvedic herbs grow only in the Himalayas. Your knowledge of local herbs is a valuable asset.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Panchakarma training in Nepal", URL: "https://www.namc.org.np"},
				{Title: "Herbal medicine formulation (YouTube)", URL: "https://www.youtube.com/results?search_query=ayurvedic+herbal+formulation"},
				{Title: "Ayurvedic pharmacopoeia of Nepal", URL: "https://www.mohp.gov.np"},
			}},
			{StepNumber: 6, Title: "Grow your practice or start an Ayurveda center", Description: "Build a successful practice with loyal patients. Open your own Ayurveda clinic or wellness center. Many patients come from abroad for Ayurvedic treatment in Nepal — especially from Europe, Russia, and North America. Write articles, make videos, or teach to share your knowledge. The global interest in natural medicine is growing and Nepal is one of the best places in the world to practice authentic Ayurveda. You carry forward a tradition that is thousands of years old and more relevant today than ever.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Opening an Ayurveda clinic in Nepal", URL: "https://www.namc.org.np"},
				{Title: "Ayurvedic tourism in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=ayurveda+tourism+nepal"},
				{Title: "Marketing your Ayurveda practice", URL: "https://www.youtube.com/results?search_query=ayurveda+marketing+nepal"},
			}},
		},
	}
}

func communityHealthWorker() careerSeed {
	return careerSeed{
		CategoryName: "Healthcare & Wellness",
		CategorySlug: "healthcare-wellness",
		CategoryIcon: "🏥",
		Title:        "Community Health Worker",
		Slug:         "community-health-worker",
		Summary:      "Community health workers provide basic healthcare and health education in villages and communities, especially where hospitals are far away.",
		Description:  "A community health worker (also called Female Community Health Volunteer - FCHV in Nepal) provides basic health services in communities. They promote maternal and child health, family planning, nutrition, immunization, and hygiene. They treat common illnesses like diarrhea, respiratory infections, and minor wounds. They also collect health data and report to health posts. Nepal's network of FCHVs is world-famous for improving health outcomes, especially reducing maternal and child mortality. Community health workers work in health posts, go door-to-door, and organize health awareness events. This career is especially important in rural areas where access to doctors is limited. It is a deeply satisfying career for people who want to help their community.",
		DailyTasks: []string{
			"Visit homes to check on pregnant women and new mothers",
			"Conduct immunization drives for children in the community",
			"Treat common illnesses like colds, diarrhea, and skin infections",
			"Educate families about hygiene, nutrition, and family planning",
			"Keep records of births, deaths, and health data",
			"Distribute vitamins, iron tablets, and basic medicines",
			"Refer serious cases to hospitals or health posts",
		},
		Skills: []string{
			"Basic medical knowledge and clinical skills",
			"Communication and health education skills",
			"Empathy and cultural sensitivity",
			"Record keeping and basic data collection",
			"Community organizing and mobilization",
			"Knowledge of maternal and child health",
			"Ability to work independently in rural settings",
		},
		SalaryMin:   150000,
		SalaryMax:   500000,
		Difficulty:  2,
		FutureProof: 75,
		EducationReq: "SLC/SEE pass minimum. Training from the Ministry of Health and Population or local health offices. Female Community Health Volunteer (FCHV) training program. ANM (Auxiliary Nurse Midwife) training through CTEVT for higher-level roles.",
		Outlook:      "Community health workers are the backbone of Nepal's rural healthcare system. The government continues to invest in community health programs. International organizations (WHO, UNICEF, USAID) support community health initiatives. The need is greatest in remote areas. The career offers deep satisfaction but limited financial rewards compared to other healthcare roles.",
		Tags:         []string{"healthcare", "community", "helping-people", "rural", "women-empowerment"},
		Resources: []resourceSeed{
			{Title: "Ministry of Health and Population Nepal", URL: "https://www.mohp.gov.np", Description: "Official health ministry resources and programs"},
			{Title: "WHO Nepal - Community Health", URL: "https://www.who.int/nepal", Description: "WHO resources for community health workers in Nepal"},
			{Title: "Merojob Healthcare Jobs", URL: "https://www.merojob.com", Description: "Find community health worker jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Complete SLC/SEE and express interest in helping your community", Description: "Finish your basic education. If you live in a rural area, you already understand the health challenges your community faces. Talk to the Female Community Health Volunteer (FCHV) in your village — ask them about their work. Read about basic health topics: maternal health, child nutrition, common diseases. Your desire to help others is the most important qualification. Everything else can be learned. If you care about your community, you already have the right heart for this work.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Ministry of Health Nepal - Community programs", URL: "https://www.mohp.gov.np"},
				{Title: "WHO - Community health worker guide", URL: "https://www.who.int/nepal"},
				{Title: "FCHV program in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=fchv+nepal+program"},
			}},
			{StepNumber: 2, Title: "Complete health worker training", Description: "Enroll in the Female Community Health Volunteer (FCHV) training program offered by the Ministry of Health. The training covers: maternal and child health, family planning, nutrition, immunization, common disease management, record keeping, and communication skills. Training is usually conducted at the local health office or district hospital. It typically takes 3-6 months. For higher-level roles, complete ANM (Auxiliary Nurse Midwife) training at a CTEVT institute (18 months). The more training you have, the more you can help your community.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "FCHV training program - Ministry of Health", URL: "https://www.mohp.gov.np"},
				{Title: "CTEVT - ANM training programs", URL: "https://www.ctevt.org.np"},
				{Title: "Basic health skills training (YouTube)", URL: "https://www.youtube.com/results?search_query=community+health+worker+training+nepal"},
			}},
			{StepNumber: 3, Title: "Start serving your community with regular visits", Description: "Go door-to-door to meet families in your assigned area. Identify pregnant women, new mothers, and children who need vaccinations. Provide health education during home visits. Hold health awareness sessions at the local health post or community building. Treat basic illnesses. Keep records of every visit. Build trust with families — they need to know you are there to help, not to judge. Trust takes time but it is the foundation of effective community health work.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Home visit techniques for health workers (YouTube)", URL: "https://www.youtube.com/results?search_query=community+health+home+visit+tips"},
				{Title: "Health education materials for communities", URL: "https://www.mohp.gov.np"},
				{Title: "Maternal and child health best practices", URL: "https://www.who.int/nepal"},
			}},
			{StepNumber: 4, Title: "Conduct immunization campaigns and health events", Description: "Organize and participate in immunization drives in your community. Coordinate with the health post for vaccine supplies. Hold health camps for specific issues: eye checkups, deworming, nutrition screening. Celebrate national health days (e.g., National Immunization Day). These events bring the community together and improve health outcomes for everyone. Each immunization you give or health talk you conduct saves lives that you may never know about — but the impact is real.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Immunization program in Nepal", URL: "https://www.mohp.gov.np"},
				{Title: "Organizing health camps in rural Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=health+camp+nepal+community"},
				{Title: "Health education communication strategies", URL: "https://www.youtube.com/results?search_query=health+education+tips+nepal"},
			}},
			{StepNumber: 5, Title: "Get additional training and take on more responsibility", Description: "Seek advanced training in specific areas: nutrition counseling, tuberculosis treatment support, HIV/AIDS awareness, family planning methods, or disaster health response. Experienced community health workers often supervise newer volunteers. Some become Health Assistants or work in rural health posts. With additional qualifications, you can advance to more specialized and better-paying roles. The government values experienced community health workers and provides pathways for growth.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Advanced health worker training programs", URL: "https://www.mohp.gov.np"},
				{Title: "Nutrition counseling training (YouTube)", URL: "https://www.youtube.com/results?search_query=nutrition+counseling+training"},
				{Title: "Supervisory skills for health workers", URL: "https://www.youtube.com/results?search_query=supervisory+skills+health+nepal"},
			}},
			{StepNumber: 6, Title: "Become a leader in community health", Description: "Experienced community health workers are respected leaders in their communities. Some go on to hold elected positions in local government (ward member, vice-chair). Others train new health workers or work with international NGOs. Some pursue higher education in public health and become public health professionals. The skills you learn — compassion, communication, organization — serve you for life. Community health workers are unsung heroes who save lives every day in Nepal's villages. You can be one of them.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Public health career paths in Nepal", URL: "https://www.mohp.gov.np"},
				{Title: "NGO opportunities for health workers (YouTube)", URL: "https://www.youtube.com/results?search_query=ngo+health+worker+nepal"},
				{Title: "Community health leadership guide", URL: "https://www.who.int/nepal"},
			}},
		},
	}
}

func dentist() careerSeed {
	return careerSeed{
		CategoryName: "Healthcare & Wellness",
		CategorySlug: "healthcare-wellness",
		CategoryIcon: "🦷",
		Title:        "Dentist",
		Slug:         "dentist",
		Summary:      "Dentists treat problems with teeth and gums. They clean teeth, fill cavities, perform surgeries, and help people maintain good oral health.",
		Description:  "A dentist diagnoses and treats problems related to teeth, gums, and the mouth. They clean teeth, fill cavities, perform root canals, extract damaged teeth, fit dentures and crowns, and treat gum disease. In Nepal, dentists work in private dental clinics, hospitals, dental colleges, and government health facilities. The demand for dental services is growing as more Nepalis understand the importance of oral health. Dental tourism is also emerging in Nepal, with patients from neighboring countries seeking affordable dental care. Becoming a dentist requires a BDS (Bachelor of Dental Surgery) degree followed by registration with the Nepal Medical Council.",
		DailyTasks: []string{
			"Examine patients' teeth and gums for problems",
			"Clean teeth and remove plaque and tartar",
			"Fill cavities with appropriate materials",
			"Perform root canal treatments",
			"Extract damaged or decayed teeth",
			"Take and interpret dental X-rays",
			"Advise patients on oral hygiene and care",
		},
		Skills: []string{
			"Manual dexterity and precision hand skills",
			"Knowledge of dental anatomy and procedures",
			"Patient communication and empathy",
			"Diagnostic skills (reading X-rays, clinical examination)",
			"Infection control and sterilization procedures",
			"Business management (for private practice)",
			"Ability to work with anxious patients",
		},
		SalaryMin:   500000,
		SalaryMax:   3000000,
		Difficulty:  4,
		FutureProof: 85,
		EducationReq: "Bachelor of Dental Surgery (BDS) from a recognized university. Must register with the Nepal Medical Council (NMC) to practice. Specialization (MDS) for higher pay and advanced procedures. Nepal Dental Association membership recommended.",
		Outlook:      "Oral health awareness is increasing in Nepal, creating more demand for dental services. The number of dental colleges in Nepal is growing. Private practice offers good income potential. Dental tourism is an emerging market. Rural areas have fewer dentists, creating opportunities for those willing to serve in underserved areas.",
		Tags:         []string{"healthcare", "medical", "doctor", "stable", "high-salary"},
		Resources: []resourceSeed{
			{Title: "Nepal Medical Council - Dentist Registration", URL: "https://www.nmc.org.np", Description: "Official registration and licensing for dentists in Nepal"},
			{Title: "Nepal Dental Association", URL: "https://www.nepaldentalassociation.org", Description: "Professional body for dentists in Nepal with resources and networking"},
			{Title: "Merojob Healthcare Jobs", URL: "https://www.merojob.com", Description: "Find dentist jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Focus on science in high school", Description: "Take biology, chemistry, and physics seriously in SEE and +2. Good grades are essential for admission to BDS programs which are competitive. Read about dentistry and oral health. Visit a dental clinic and ask the dentist about their work. Shadowing a dentist for a day can confirm your interest. Dentistry requires good hand-eye coordination. If you enjoy detailed, precise work with your hands, this could be the right path for you.", Duration: "2 years", Links: []roadmapLink{
				{Title: "Nepal Medical Council - BDS colleges", URL: "https://www.nmc.org.np"},
				{Title: "Career in dentistry overview (YouTube)", URL: "https://www.youtube.com/results?search_query=dentistry+career+nepal"},
				{Title: "BDS entrance exam preparation Nepal", URL: "https://www.youtube.com/results?search_query=bds+entrance+nepal"},
			}},
			{StepNumber: 2, Title: "Complete BDS (Bachelor of Dental Surgery)", Description: "Enroll in a BDS program at a recognized dental college in Nepal (Kathmandu University, BP Koirala Institute of Health Sciences, or KU). The program takes 5 years including internship. Study dental anatomy, oral pathology, periodontology, orthodontics, prosthodontics, and oral surgery. Practice on models before treating real patients during internship. BDS is challenging but the skills you learn will serve patients for your entire career. Your first filling on a real patient is a milestone you will never forget.", Duration: "5 years", Links: []roadmapLink{
				{Title: "BP Koirala Institute of Health Sciences - BDS", URL: "https://www.bpkihs.edu"},
				{Title: "Kathmandu University - Dental program", URL: "https://www.ku.edu.np"},
				{Title: "Free dental education resources (YouTube)", URL: "https://www.youtube.com/results?search_query=dental+education+for+bds"},
			}},
			{StepNumber: 3, Title: "Register with Nepal Medical Council and get licensed", Description: "After completing BDS and internship, register with the Nepal Medical Council. Submit your degree, internship certificate, and other documents. Pass the NMC licensing exam. Once registered, you can practice dentistry legally in Nepal. Registration must be renewed periodically. The NMC sets the standards for dental practice and ensures patient safety. Your license is your professional identity. Protect it by following ethical practices and continuing your education.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Nepal Medical Council - Dentist registration", URL: "https://www.nmc.org.np"},
				{Title: "NMC licensing exam for dentists (YouTube)", URL: "https://www.youtube.com/results?search_query=nmc+dentist+license+exam+nepal"},
				{Title: "Nepal Dental Association - Professional resources", URL: "https://www.nepaldentalassociation.org"},
			}},
			{StepNumber: 4, Title: "Start practicing and building experience", Description: "Join a dental hospital, clinic, or open your own practice. Many new dentists work as assistants in established clinics to build experience. Consider working in a rural area where there is less competition and greater need. Build your skills in all areas of general dentistry. Develop good relationships with patients — a friendly, reassuring manner is as important as technical skill. A patient who trusts you will come back for years and refer their family. Your reputation is your most valuable asset.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Find dentist jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Opening a dental clinic in Nepal", URL: "https://www.nepaldentalassociation.org"},
				{Title: "Dental practice management tips (YouTube)", URL: "https://www.youtube.com/results?search_query=dental+practice+management+nepal"},
			}},
			{StepNumber: 5, Title: "Specialize with MDS (Master of Dental Surgery)", Description: "After 2-3 years of general practice, consider specializing. Popular specializations: orthodontics (braces), prosthodontics (crowns, bridges, dentures), endodontics (root canals), oral surgery, periodontics (gums), or pediatric dentistry. MDS takes 3 years. Specialists earn significantly more and have more satisfied patients. Specialization also means you focus on what you enjoy most. If you love the precision of root canals, become an endodontist. If you enjoy transforming smiles, become an orthodontist.",
			Duration: "3 years", Links: []roadmapLink{
				{Title: "MDS specialization options in Nepal", URL: "https://www.nmc.org.np"},
				{Title: "Orthodontics training and career (YouTube)", URL: "https://www.youtube.com/results?search_query=orthodontics+career+nepal"},
				{Title: "Oral surgery specialization guide", URL: "https://www.youtube.com/results?search_query=oral+surgery+career+nepal"},
			}},
			{StepNumber: 6, Title: "Grow your practice or become a dental educator", Description: "Experienced dentists can grow their own clinic chain, work as consultants for dental hospitals, or become professors at dental colleges. Some dentists participate in dental tourism, treating international patients who come to Nepal for affordable dental care. Others work with NGOs providing dental camps in rural areas. Leadership roles in the Nepal Dental Association allow you to shape the future of dentistry in Nepal. A dentist's work changes lives — a healthy smile gives people confidence and improves their quality of life.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Building a multi-specialty dental clinic", URL: "https://www.nepaldentalassociation.org"},
				{Title: "Dental tourism opportunities Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=dental+tourism+nepal"},
				{Title: "Academic career in dentistry", URL: "https://www.ku.edu.np"},
			}},
		},
	}
}

func radiographer() careerSeed {
	return careerSeed{
		CategoryName: "Healthcare & Wellness",
		CategorySlug: "healthcare-wellness",
		CategoryIcon: "📷",
		Title:        "Radiographer",
		Slug:         "radiographer",
		Summary:      "Radiographers operate X-ray, CT scan, MRI, and ultrasound machines to create medical images that help doctors diagnose diseases.",
		Description:  "A radiographer (also called a radiologic technologist) operates medical imaging equipment to create images of the inside of a patient's body. These images help doctors diagnose fractures, tumors, infections, and other medical conditions. Radiographers work with X-ray machines, CT (computed tomography) scanners, MRI (magnetic resonance imaging) machines, ultrasound equipment, and mammography machines. In Nepal, radiographers work in hospitals, diagnostic centers, and clinics. The field is growing as more medical facilities install advanced imaging equipment. Radiographers must follow strict safety protocols to protect themselves and patients from radiation. It is a technical job that combines patient care with technology.",
		DailyTasks: []string{
			"Prepare patients for imaging procedures and explain the process",
			"Position patients correctly for X-rays and scans",
			"Operate X-ray, CT, MRI, and ultrasound equipment",
			"Adjust radiation exposure settings for each patient",
			"Ensure radiation safety protocols are followed",
			"Review images for quality before sending to radiologists",
			"Maintain imaging equipment and report problems",
		},
		Skills: []string{
			"Operation of X-ray, CT, MRI, and ultrasound machines",
			"Patient positioning and immobilization techniques",
			"Knowledge of human anatomy and body positioning",
			"Radiation safety and protection practices",
			"Attention to detail and image quality assessment",
			"Patient communication and reassurance",
			"Basic computer skills for digital imaging systems",
		},
		SalaryMin:   300000,
		SalaryMax:   1000000,
		Difficulty:  3,
		FutureProof: 78,
		EducationReq: "Diploma in Radiography (3 years) from CTEVT or Bachelor's in Radiography (B.Sc. MIT) from PU/KU. Must register with the Nepal Health Professional Council. Additional certifications for CT, MRI, and mammography are valuable.",
		Outlook:      "The demand for radiographers in Nepal is growing as more hospitals and diagnostic centers open. New imaging technologies (digital X-ray, advanced CT/MRI) create opportunities for trained professionals. There is a shortage of qualified radiographers in rural areas. The field offers stable employment with potential for advancement into specialized areas like interventional radiology or nuclear medicine.",
		Tags:         []string{"healthcare", "medical", "technology", "imaging", "stable"},
		Resources: []resourceSeed{
			{Title: "CTEVT Nepal - Radiography programs", URL: "https://www.ctevt.org.np", Description: "Diploma in Radiography and related programs"},
			{Title: "Nepal Health Professional Council", URL: "https://www.nhpc.gov.np", Description: "Regulatory body for radiographers and other health professionals"},
			{Title: "Merojob Healthcare Jobs", URL: "https://www.merojob.com", Description: "Find radiographer jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Study science and understand what radiography involves", Description: "Focus on physics, biology, and chemistry in SEE and +2 level. Physics understanding is especially important for radiography (X-rays, radiation physics). Visit a diagnostic center and ask to see the X-ray room. Talk to a radiographer about their daily work. Make sure you are comfortable with technology and patient interaction. Radiography involves close contact with patients who may be in pain. Compassion is as important as technical skill.", Duration: "1-2 years", Links: []roadmapLink{
				{Title: "CTEVT - Radiography career info", URL: "https://www.ctevt.org.np"},
				{Title: "What does a radiographer do? (YouTube)", URL: "https://www.youtube.com/results?search_query=radiography+career+nepal"},
				{Title: "Physics basics for radiography", URL: "https://www.youtube.com/results?search_query=radiation+physics+for+beginners"},
			}},
			{StepNumber: 2, Title: "Complete radiography training (Diploma or Bachelor's)", Description: "Enroll in a Diploma in Radiography (3 years) at a CTEVT-affiliated institute or a B.Sc. MIT (Medical Imaging Technology) at a university. Study anatomy, radiographic techniques, radiation physics, darkroom procedures (for traditional film), and digital imaging. Practical training is a major part of the program — you will practice on positioning phantoms and eventually real patients during clinical placements. Learn each projection carefully. Correct positioning is the difference between a diagnostic image and a useless one.", Duration: "3-4 years", Links: []roadmapLink{
				{Title: "CTEVT - Affiliated radiography colleges", URL: "https://www.ctevt.org.np"},
				{Title: "Pokhara University - B.Sc. MIT program", URL: "https://www.pu.edu.np"},
				{Title: "Radiography positioning guide (YouTube)", URL: "https://www.youtube.com/results?search_query=radiography+positioning+guide"},
			}},
			{StepNumber: 3, Title: "Get licensed and start working", Description: "Register with the Nepal Health Professional Council after completing your training. Apply for positions in hospitals (both government and private), diagnostic centers, and clinics. Entry-level radiographers typically start with basic X-ray work. Build speed and accuracy. Learn to handle different types of patients: children, elderly, trauma patients. Develop your technique for getting high-quality images with minimal radiation exposure. The ALARA principle — As Low As Reasonably Achievable — guides all radiation work.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "NHPC - Radiographer registration", URL: "https://www.nhpc.gov.np"},
				{Title: "Find radiographer jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Radiation safety for radiographers (YouTube)", URL: "https://www.youtube.com/results?search_query=radiation+safety+for+radiographers"},
			}},
			{StepNumber: 4, Title: "Learn advanced imaging modalities", Description: "Expand your skills to advanced modalities: CT scanning, MRI, ultrasound, and mammography. Each requires additional training and certification. CT and MRI technologists earn more than general radiographers. Many hospitals will pay for your advanced training if you commit to working with them. Learn digital radiography systems and Picture Archiving and Communication Systems (PACS). The more modalities you can operate, the more valuable you are to employers.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "CT scan training for radiographers (YouTube)", URL: "https://www.youtube.com/results?search_query=ct+scan+training+for+beginners"},
				{Title: "MRI safety and operation training", URL: "https://www.youtube.com/results?search_query=mri+training+for+radiographers"},
				{Title: "Ultrasound imaging basics", URL: "https://www.youtube.com/results?search_query=ultrasound+imaging+for+beginners"},
			}},
			{StepNumber: 5, Title: "Specialize or become a senior radiographer", Description: "After 3-5 years, choose a path: specialize in one modality (CT, MRI, interventional radiology, or nuclear medicine), become a chief radiographer (supervising other staff), or move into application training (teaching others how to use new equipment). Interventional radiology technologists assist doctors during minimally invasive procedures. This is a high-skill, high-reward specialty. Specialization brings higher pay and more respect in the medical community.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Interventional radiology technologist career", URL: "https://www.youtube.com/results?search_query=interventional+radiology+tech"},
				{Title: "Nuclear medicine technology training", URL: "https://www.youtube.com/results?search_query=nuclear+medicine+technology"},
				{Title: "Chief radiographer role and responsibilities", URL: "https://www.youtube.com/results?search_query=chief+radiographer+role"},
			}},
			{StepNumber: 6, Title: "Advance into management, education, or sales", Description: "Senior radiographers can become department managers, overseeing all imaging services in a hospital. Some teach at CTEVT institutes and universities, training the next generation of radiographers. Others work for medical equipment companies as application specialists or sales representatives — these roles pay very well. The imaging technology field is always evolving (AI-powered imaging, portable ultrasound, 3D mammography). Staying current with technology keeps your skills valuable and your career growing.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Radiology department management (YouTube)", URL: "https://www.youtube.com/results?search_query=radiology+department+management"},
				{Title: "Teaching career in radiography", URL: "https://www.ctevt.org.np"},
				{Title: "Medical imaging equipment sales career", URL: "https://www.youtube.com/results?search_query=medical+equipment+sales+career"},
			}},
		},
	}
}

func physiotherapist() careerSeed {
	return careerSeed{
		CategoryName: "Healthcare & Wellness",
		CategorySlug: "healthcare-wellness",
		CategoryIcon: "💪",
		Title:        "Physiotherapist",
		Slug:         "physiotherapist",
		Summary:      "Physiotherapists help people recover from injuries, surgeries, and chronic pain. They use exercises, massages, and other techniques to restore movement and function.",
		Description:  "A physiotherapist (physical therapist) helps patients recover from injuries, surgeries, and conditions that affect movement. They use exercises, manual therapy (massage and mobilization), electrotherapy, heat/cold therapy, and education to help patients regain strength and mobility. In Nepal, physiotherapy is a growing field with demand in hospitals, rehabilitation centers, sports clinics, orthopedic clinics, neurology departments, and geriatric care. The number of physiotherapy training programs in Nepal has increased significantly. Physiotherapists are needed for sports injuries, stroke rehabilitation, back pain, arthritis, post-surgical recovery, and children with developmental delays.",
		DailyTasks: []string{
			"Assess patients' movement, strength, and range of motion",
			"Develop treatment plans with specific exercises",
			"Guide patients through therapeutic exercises",
			"Perform manual therapy (massage, joint mobilization)",
			"Use modalities like ultrasound, TENS, heat, and ice",
			"Educate patients on home exercises and injury prevention",
			"Track patient progress and adjust treatments",
		},
		Skills: []string{
			"Knowledge of human anatomy, physiology, and biomechanics",
			"Assessment and diagnostic skills for movement problems",
			"Therapeutic exercise prescription and guidance",
			"Manual therapy techniques (massage, mobilization)",
			"Patient education and motivation",
			"Empathy and communication with patients in pain",
			"Treatment planning and documentation",
		},
		SalaryMin:   250000,
		SalaryMax:   900000,
		Difficulty:  3,
		FutureProof: 80,
		EducationReq: "Bachelor of Physiotherapy (BPT/B.Sc. PT) from recognized university (KU, PU, BPKIHS). Must register with Nepal Health Professional Council. Specialty certifications (sports, neuro, pediatric) for advanced practice.",
		Outlook:      "Physiotherapy demand in Nepal is growing rapidly due to increased awareness, sports participation, aging population, and rising rates of lifestyle diseases (back pain, diabetes, heart disease). More hospitals are establishing physiotherapy departments. Sports physiotherapy is emerging with professional sports. Neurological rehabilitation (stroke, spinal cord injury) is an underserved area with high need.",
		Tags:         []string{"healthcare", "rehabilitation", "exercise", "helping-people", "growing"},
		Resources: []resourceSeed{
			{Title: "Nepal Health Professional Council - PT Registration", URL: "https://www.nhpc.gov.np", Description: "Registration and licensing for physiotherapists in Nepal"},
			{Title: "Nepal Physiotherapy Association", URL: "https://www.nepalphysio.org", Description: "Professional body for physiotherapists with resources and networking"},
			{Title: "Merojob Healthcare Jobs", URL: "https://www.merojob.com", Description: "Find physiotherapist jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Study biology and understand how the body moves", Description: "Focus on biology (especially human anatomy) in SEE and +2 science stream. Read about how muscles, bones, and nerves work together to create movement. If you play sports, pay attention to how your body moves and how injuries happen. Talk to a physiotherapist if you know one. Ask them what they find rewarding about their work. Physiotherapy is perfect for people who enjoy science AND want to help people directly. You get to use your hands and your brain every day.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Nepal Physiotherapy Association - Career info", URL: "https://www.nepalphysio.org"},
				{Title: "Human anatomy basics (YouTube)", URL: "https://www.youtube.com/results?search_query=human+anatomy+for+beginners"},
				{Title: "What is physiotherapy? (YouTube)", URL: "https://www.youtube.com/results?search_query=physiotherapy+career+nepal"},
			}},
			{StepNumber: 2, Title: "Complete a Bachelor's in Physiotherapy", Description: "Enroll in a BPT or B.Sc. PT program at a recognized university: Kathmandu University, Pokhara University, BP Koirala Institute of Health Sciences, or Purbanchal University. The program takes 4 years plus compulsory internship. Study anatomy, physiology, biomechanics, exercise therapy, electrotherapy, orthopedics, neurology, cardiopulmonary, and community physiotherapy. Clinical placements in hospitals give you hands-on experience with real patients. This is where you learn to apply your knowledge. The subjects are challenging but every class prepares you to help someone recover.", Duration: "4-5 years", Links: []roadmapLink{
				{Title: "Kathmandu University - Physiotherapy program", URL: "https://www.ku.edu.np"},
				{Title: "Pokhara University - B.Sc. PT", URL: "https://www.pu.edu.np"},
				{Title: "Free physiotherapy study resources (YouTube)", URL: "https://www.youtube.com/results?search_query=physiotherapy+study+nepal"},
			}},
			{StepNumber: 3, Title: "Register with NHPC and start your career", Description: "Register with the Nepal Health Professional Council to get your license. Apply for positions in hospitals, rehabilitation centers, or private physiotherapy clinics. Entry-level physiotherapists typically treat a mix of orthopedic (back pain, joint pain), neurological (stroke, paralysis), and sports injury patients. Build your clinical reasoning skills. Learn to listen to patients — they often tell you exactly what is wrong if you ask the right questions. Document your treatments carefully. Good records protect you and help track patient progress.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "NHPC - Physiotherapist registration", URL: "https://www.nhpc.gov.np"},
				{Title: "Find physiotherapy jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Clinical reasoning in physiotherapy (YouTube)", URL: "https://www.youtube.com/results?search_query=clinical+reasoning+physiotherapy"},
			}},
			{StepNumber: 4, Title: "Develop skills in your area of interest", Description: "Take continuing education courses in areas that interest you: sports physiotherapy, neurological rehabilitation, pediatric physiotherapy, orthopedic manual therapy, or cardiopulmonary physiotherapy. Attend workshops and conferences organized by the Nepal Physiotherapy Association. The more specialized your skills, the more you can help patients with complex conditions. Consider learning dry needling, Kinesio taping, or advanced manual therapy techniques. Each new skill makes you a more effective therapist.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Nepal Physiotherapy Association - Workshops", URL: "https://www.nepalphysio.org"},
				{Title: "Sports physiotherapy training (YouTube)", URL: "https://www.youtube.com/results?search_query=sports+physiotherapy+training"},
				{Title: "Stroke rehabilitation physiotherapy", URL: "https://www.youtube.com/results?search_query=stroke+rehabilitation+physiotherapy"},
			}},
			{StepNumber: 5, Title: "Specialize and become an expert", Description: "After 3-5 years, pursue a Master's in Physiotherapy (MPT) with a specialty like orthopedics, neurology, sports, or pediatrics. Specialized physiotherapists are in higher demand and can charge more. Some physiotherapists open their own clinic. Others work with sports teams — Nepal's national cricket and football teams need physiotherapists. Neurological physiotherapy (helping stroke and spinal cord injury patients) is deeply rewarding and always needed. Find the area that makes you excited to go to work every day.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "MPT specialization options in Nepal", URL: "https://www.nepalphysio.org"},
				{Title: "Sports physiotherapy in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=sports+physiotherapy+nepal"},
				{Title: "Opening your own physiotherapy clinic", URL: "https://www.youtube.com/results?search_query=start+physiotherapy+clinic+nepal"},
			}},
			{StepNumber: 6, Title: "Grow into leadership, teaching, or entrepreneurship", Description: "Experienced physiotherapists can become department heads, clinical specialists, or professors at physiotherapy colleges. Some work with international NGOs on rehabilitation projects. Others develop home health services, visiting patients who cannot travel to clinics. The World Health Organization recognizes the importance of rehabilitation and Nepal's healthcare system is expanding physiotherapy services. As a physiotherapist, you help people do what they love again — walk, run, play, work, live without pain. That is a powerful gift to give.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Physiotherapy management career (YouTube)", URL: "https://www.youtube.com/results?search_query=physiotherapy+management+nepal"},
				{Title: "Home health physiotherapy services", URL: "https://www.nepalphysio.org"},
				{Title: "Teaching career in physiotherapy", URL: "https://www.ku.edu.np"},
			}},
		},
	}
}

func yogaInstructor() careerSeed {
	return careerSeed{
		CategoryName: "Healthcare & Wellness",
		CategorySlug: "healthcare-wellness",
		CategoryIcon: "🧘",
		Title:        "Yoga Instructor",
		Slug:         "yoga-instructor",
		Summary:      "Yoga instructors teach yoga postures, breathing exercises, and meditation to help people improve their physical and mental health.",
		Description:  "A yoga instructor leads individuals and groups in yoga practice. They teach physical postures (asanas), breathing techniques (pranayama), and meditation. Yoga originated in South Asia thousands of years ago and Nepal is part of this tradition. Yoga is extremely popular worldwide, and Nepal has become a global destination for yoga teacher training and retreats. Yoga instructors in Nepal can work in studios, hotels, wellness centers, or independently. Many tourists come to Nepal specifically for yoga retreats. The Yoga Certification Board of Nepal provides certification standards. This career combines physical activity, spiritual practice, and helping others. It is especially popular among those who value holistic health.",
		DailyTasks: []string{
			"Plan and lead yoga classes for groups of students",
			"Demonstrate proper alignment in yoga postures",
			"Provide hands-on adjustments and modifications",
			"Teach breathing exercises and meditation techniques",
			"Offer one-on-one sessions for private clients",
			"Clean and prepare the yoga studio space",
			"Continue personal practice to improve own skills",
		},
		Skills: []string{
			"Deep knowledge of yoga asanas, pranayama, and meditation",
			"Ability to demonstrate and explain poses clearly",
			"Public speaking and class management",
			"Anatomy knowledge for safe alignment",
			"Patience and empathy with students of all levels",
			"Therapeutic knowledge (yoga for specific conditions)",
			"English language (for international students)",
		},
		SalaryMin:   200000,
		SalaryMax:   1000000,
		Difficulty:  2,
		FutureProof: 68,
		EducationReq: "200-hour Yoga Teacher Training (YTT) certificate from a Yoga Certification Board Nepal registered school. 500-hour YTT for advanced teaching. Anatomy and physiology knowledge helpful. English language important for teaching international students.",
		Outlook:      "Yoga is growing globally and Nepal is a top destination for yoga tourism. Yoga retreats around Pokhara and Kathmandu attract thousands of visitors. Online yoga teaching is also growing. Competition exists but skilled and certified instructors are always in demand. The Nepal government is supporting yoga tourism as part of its tourism strategy.",
		Tags:         []string{"wellness", "health", "fitness", "tourism", "spiritual"},
		Resources: []resourceSeed{
			{Title: "Yoga Certification Board Nepal", URL: "https://www.yogacertificationnepal.com", Description: "Official yoga teacher certification body in Nepal"},
			{Title: "Yoga Retreat Nepal - Teacher Training", URL: "https://www.yogainnepal.com", Description: "Yoga teacher training programs and resources"},
			{Title: "Merojob Wellness Jobs", URL: "https://www.merojob.com", Description: "Find yoga instructor jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Start your personal yoga practice", Description: "Begin practicing yoga regularly. Start with free YouTube classes or join a local yoga studio. Practice at least 3-4 times per week. Learn the basic postures: sun salutation, standing poses, seated poses, backbends, twists, and inversions. Learn to breathe properly during practice. Keep a journal of your practice. Yoga is experiential — you cannot teach what you have not experienced. Your own body will be your first and most important teacher. Consistency matters more than intensity. Even 15 minutes a day is better than 2 hours once a week.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Yoga for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=yoga+for+beginners+2024"},
				{Title: "Yoga studios in Kathmandu/Pokhara", URL: "https://www.yogainnepal.com"},
				{Title: "Sun salutation guide (YouTube)", URL: "https://www.youtube.com/results?search_query=sun+salutation+guide"},
			}},
			{StepNumber: 2, Title: "Complete a 200-hour Yoga Teacher Training (YTT)", Description: "Enroll in a registered 200-hour YTT program. Nepal has many excellent YTT programs in Kathmandu, Pokhara, and Chitwan. The training covers asanas (poses), pranayama (breathing), meditation, anatomy, philosophy, teaching methodology, and practice teaching. Most programs are 3-4 weeks intensive or spread over several months. You will learn not just how to do yoga, but how to guide others. You will also deepen your own practice significantly. Choose a school registered with the Yoga Certification Board Nepal for quality assurance.",
			Duration: "1-4 months", Links: []roadmapLink{
				{Title: "Yoga Certification Board Nepal - Registered schools", URL: "https://www.yogacertificationnepal.com"},
				{Title: "200-hour YTT in Pokhara (YouTube)", URL: "https://www.youtube.com/results?search_query=200+hour+ytt+pokhara+nepal"},
				{Title: "Yoga teacher training experience (YouTube)", URL: "https://www.youtube.com/results?search_query=yoga+teacher+training+nepal+experience"},
			}},
			{StepNumber: 3, Title: "Start teaching — practice teaching any chance you get", Description: "Offer free classes to friends, family, and community groups. Teach at a local community center or park. Practice your cueing (telling students what to do). Learn to observe students and offer adjustments. Teaching is a skill separate from practicing — you develop it through repetition. Record yourself teaching and watch it back. Ask for feedback. Your first class will be nerve-wracking but it gets easier. Every great yoga teacher was once a nervous beginner. Keep teaching and you will improve naturally.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "How to teach your first yoga class (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+teach+your+first+yoga+class"},
				{Title: "Yoga class sequencing tips", URL: "https://www.youtube.com/results?search_query=yoga+class+sequencing+for+teachers"},
				{Title: "Teaching yoga to beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=teaching+yoga+to+beginners+tips"},
			}},
			{StepNumber: 4, Title: "Find teaching opportunities in Nepal's yoga tourism industry", Description: "Apply to teach at yoga studios, hotels, and wellness centers in Kathmandu, Pokhara, and Chitwan. The yoga tourism industry in Nepal employs many instructors, especially during tourist season (October-May). Offer to substitute teach at studios. Build relationships with studio owners. Create a schedule of regular classes. Many instructors also offer private sessions to tourists. Your ability to speak English well is very important for teaching international students. A friendly, welcoming manner matters as much as your yoga knowledge.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Find yoga teaching jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Yoga retreats in Nepal - Teaching opportunities", URL: "https://www.yogainnepal.com"},
				{Title: "Building a yoga teaching schedule (YouTube)", URL: "https://www.youtube.com/results?search_query=build+yoga+teaching+schedule"},
			}},
			{StepNumber: 5, Title: "Specialize in a style or therapeutic area", Description: "After your first year of teaching, specialize: Hatha yoga, Vinyasa flow, Yin yoga, Ashtanga, or therapeutic yoga (yoga for back pain, anxiety, prenatal, seniors). Get additional certifications in your chosen specialty. A 500-hour YTT deepens your knowledge significantly. Consider learning yoga therapy — how to use yoga to help people with specific health conditions. Specialized instructors earn more and have more dedicated students. Find the style that matches your personality and teaching strengths.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Advanced yoga teacher training (500-hour)", URL: "https://www.yogacertificationnepal.com"},
				{Title: "Yoga therapy training (YouTube)", URL: "https://www.youtube.com/results?search_query=yoga+therapy+training+online"},
				{Title: "Specialized yoga styles guide", URL: "https://www.youtube.com/results?search_query=yoga+styles+comparison"},
			}},
			{StepNumber: 6, Title: "Build your brand — studio, retreats, or online", Description: "Experienced yoga instructors can open their own studio, lead yoga retreats (in Pokhara, Chitwan, or Lumbini), teach online classes to international students, or combine yoga with other wellness services (Ayurveda, massage, meditation). Create a website and social media presence. Build a community of students who love your teaching. YouTube and Instagram can help you reach students worldwide. Yoga is not just a job in Nepal — it is part of the country's heritage. Sharing yoga with others is sharing peace, health, and happiness.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Opening a yoga studio in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=open+yoga+studio+nepal"},
				{Title: "Leading yoga retreats in Nepal", URL: "https://www.yogainnepal.com"},
				{Title: "Teaching yoga online (YouTube)", URL: "https://www.youtube.com/results?search_query=teach+yoga+online+as+a+business"},
			}},
		},
	}
}

func schoolTeacher() careerSeed {
	return careerSeed{
		CategoryName: "Education",
		CategorySlug: "education",
		CategoryIcon: "📖",
		Title:        "School Teacher (Secondary)",
		Slug:         "school-teacher",
		Summary:      "Secondary school teachers teach specific subjects like math, science, English, or social studies to students in grades 8-12.",
		Description:  "A secondary school teacher teaches specific subjects to students from grade 8 to 12. They prepare lessons, give assignments, grade work, and help students prepare for the SEE (Secondary Education Examination) and +2 board exams. In Nepal, secondary teachers can work in government schools (granted by the Public Service Commission / Loksewa), private schools, or community schools. The government has made teaching a respected and stable career with good benefits. The Teacher Service Commission (TSC) manages recruitment for government school teachers. Subjects in demand include mathematics, science, English, computer science, and social studies. Teaching in rural areas often comes with additional allowances.",
		DailyTasks: []string{
			"Prepare lesson plans for classes according to the curriculum",
			"Teach subject content to students in engaging ways",
			"Grade assignments, tests, and exam papers",
			"Help struggling students one-on-one after class",
			"Manage classroom behavior and discipline",
			"Attend staff meetings and professional development",
			"Communicate with parents about student progress",
		},
		Skills: []string{
			"In-depth knowledge of your subject area",
			"Lesson planning and curriculum delivery",
			"Classroom management and student engagement",
			"Communication with students, parents, and staff",
			"Assessment and grading skills",
			"Patience and empathy with adolescent students",
			"Creativity in making lessons interesting",
		},
		SalaryMin:   250000,
		SalaryMax:   800000,
		Difficulty:  3,
		FutureProof: 72,
		EducationReq: "Bachelor's degree in Education (B.Ed.) with specialization in the subject you want to teach. Master's degree preferred. Must pass Teacher Service Commission (TSC) exam for government school positions. Private schools may have more flexible requirements.",
		Outlook:      "Teaching is a stable career in Nepal. Government school teachers have job security and benefits through the TSC. There is a shortage of qualified teachers in rural areas and for subjects like math, science, and English. Private schools offer competitive salaries for experienced teachers. The demand for quality education is growing in Nepal.",
		Tags:         []string{"education", "teaching", "stable", "government-job", "helping-people"},
		Resources: []resourceSeed{
			{Title: "Teacher Service Commission Nepal", URL: "https://www.tsc.gov.np", Description: "Official body for teacher recruitment and licensing in Nepal"},
			{Title: "Ministry of Education Nepal", URL: "https://www.moe.gov.np", Description: "Education policy and resources for teachers"},
			{Title: "Merojob Education Jobs", URL: "https://www.merojob.com", Description: "Find teaching jobs in Nepal"},
			{Title: "Kullabs - Free Teaching Resources", URL: "https://www.kullabs.com", Description: "Free lesson plans and study materials for Nepali teachers"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Decide which subject you want to teach", Description: "Pick a subject you love and understand well: mathematics, science, English, Nepali, social studies, or computer science. The best teachers teach subjects they are passionate about. Focus on that subject in your +2 and Bachelor's studies. Read widely in your subject. The more you know, the more you can share. Good teachers never stop learning their subject. Your enthusiasm for the subject will inspire your students.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Teacher Service Commission - Subject requirements", URL: "https://www.tsc.gov.np"},
				{Title: "Teaching career overview (YouTube)", URL: "https://www.youtube.com/results?search_query=teacher+career+nepal+guide"},
				{Title: "Kullabs - Free study resources for teachers", URL: "https://www.kullabs.com"},
			}},
			{StepNumber: 2, Title: "Complete a Bachelor's in Education (B.Ed.)", Description: "Enroll in a B.Ed. program at a Nepali university with a major in your chosen subject. Study educational psychology, teaching methods, curriculum development, and assessment techniques. Do practice teaching in real schools during your program — this is where you learn the most. Build relationships with mentor teachers who can guide you. A B.Ed. is required for government school teaching positions through the TSC. If you already have a Bachelor's in another field, you can do a one-year B.Ed. program.",
			Duration: "3-4 years", Links: []roadmapLink{
				{Title: "Tribhuvan University - B.Ed. programs", URL: "https://www.tu.edu.np"},
				{Title: "Kathmandu University - School of Education", URL: "https://www.ku.edu.np"},
				{Title: "Teaching methods and strategies (YouTube)", URL: "https://www.youtube.com/results?search_query=teaching+methods+for+secondary+teachers"},
			}},
			{StepNumber: 3, Title: "Pass the Teacher Service Commission (TSC) exam", Description: "To work in government schools, you must pass the TSC licensing exam. The exam tests your subject knowledge, teaching methodology, and general knowledge. Prepare for 3-6 months using past papers and study guides. The competition is strong but passing gives you a government teaching position with stable salary, pension, and benefits. Private school teaching does not require the TSC but pays less and has less job security. Many teachers start in private schools while preparing for the TSC.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "TSC exam preparation resources", URL: "https://www.tsc.gov.np"},
				{Title: "TSC exam past papers and study tips (YouTube)", URL: "https://www.youtube.com/results?search_query=tsc+exam+nepal+preparation"},
				{Title: "Government teaching vs private school comparison", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 4, Title: "Start teaching at a school", Description: "Once licensed, you can be placed in a government school or apply to private schools. Your first year will be the hardest — you will spend evenings planning lessons and grading papers. Every teacher goes through this. Ask experienced teachers for advice. Build good relationships with your students — they learn better from teachers they like and respect. Be firm but fair in classroom management. Your first batch of students will always hold a special place in your heart. The first time a student says 'I understand now!' — that feeling makes everything worthwhile.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Find teaching jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Tips for new teachers (YouTube)", URL: "https://www.youtube.com/results?search_query=new+teacher+tips+nepal"},
				{Title: "Classroom management strategies", URL: "https://www.youtube.com/results?search_query=classroom+management+for+secondary+teachers"},
			}},
			{StepNumber: 5, Title: "Pursue a Master's in Education (M.Ed.)", Description: "An M.Ed. degree increases your salary and opens doors to leadership roles: head teacher, curriculum developer, or school administrator. Many teachers pursue part-time Master's degrees while working. Choose a specialization: educational leadership, curriculum and instruction, or subject-specific pedagogy. Research skills learned in M.Ed. programs help you become a more effective teacher. Higher qualifications also improve your ranking in the TSC system for promotions and transfers.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "M.Ed. programs in Nepal", URL: "https://www.tu.edu.np"},
				{Title: "Educational leadership specialization (YouTube)", URL: "https://www.youtube.com/results?search_query=educational+leadership+nepal"},
				{Title: "Career advancement for teachers in Nepal", URL: "https://www.tsc.gov.np"},
			}},
			{StepNumber: 6, Title: "Move into leadership or specialized roles", Description: "Experienced teachers can become head teachers, school principals, or education officers in the Ministry of Education. Some become teacher trainers, curriculum designers, or textbook authors. Others work for NGOs in educational development. The TSC offers promotion paths based on experience and additional qualifications. Teaching in Nepal is not just a job — it is shaping the next generation of Nepali citizens. A good teacher changes lives forever and the influence of a great teacher never fades.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "School principal career path Nepal", URL: "https://www.tsc.gov.np"},
				{Title: "Teacher training and development careers", URL: "https://www.moe.gov.np"},
				{Title: "Education policy and administration careers", URL: "https://www.psc.gov.np"},
			}},
		},
	}
}

func universityProfessor() careerSeed {
	return careerSeed{
		CategoryName: "Education",
		CategorySlug: "education",
		CategoryIcon: "🎓",
		Title:        "University Professor",
		Slug:         "university-professor",
		Summary:      "University professors teach and mentor students at colleges and universities. They also conduct research, publish papers, and contribute to their academic field.",
		Description:  "A university professor teaches undergraduate and graduate students in a specific academic discipline. They deliver lectures, design courses, supervise student research, conduct their own research, publish academic papers, and serve on university committees. In Nepal, professors work at universities like Tribhuvan University, Kathmandu University, Pokhara University, Purbanchal University, and others. The path to becoming a professor requires a Master's degree (at minimum) and usually a PhD. The University Grants Commission (UGC) Nepal oversees higher education standards. Professors are respected members of society and play a key role in developing Nepal's educated workforce. Academic careers offer intellectual freedom, prestige, and the opportunity to shape young minds.",
		DailyTasks: []string{
			"Prepare and deliver lectures to university students",
			"Design course syllabi and select textbooks",
			"Grade assignments, exams, and research papers",
			"Advise and mentor students on academic and career matters",
			"Conduct research and write academic papers",
			"Attend faculty meetings and academic committee work",
			"Supervise graduate student theses and dissertations",
		},
		Skills: []string{
			"Expert-level knowledge in your academic discipline",
			"Research methodology and academic writing",
			"Public speaking and lecture delivery",
			"Student mentoring and academic advising",
			"Curriculum design and assessment development",
			"Critical thinking and analytical skills",
			"Time management balancing teaching, research, and service",
		},
		SalaryMin:   500000,
		SalaryMax:   2000000,
		Difficulty:  5,
		FutureProof: 75,
		EducationReq: "Master's degree minimum (PhD strongly preferred) in the relevant field from a recognized university. Must publish research in peer-reviewed journals. University Grants Commission (UGC) Nepal sets minimum qualification standards. Assistant Professor requires PhD or equivalent publications.",
		Outlook:      "Higher education in Nepal is expanding with new universities and colleges opening. The demand for qualified PhD holders exceeds supply. The government is investing in higher education and research. Professors in Nepal also consult for government, industry, and international organizations. Competition for permanent positions is strong.",
		Tags:         []string{"education", "academia", "prestigious", "research", "mentoring"},
		Resources: []resourceSeed{
			{Title: "University Grants Commission Nepal", URL: "https://www.ugc.gov.np", Description: "Regulatory body for higher education and university funding"},
			{Title: "Tribhuvan University", URL: "https://www.tu.edu.np", Description: "Nepal's largest university with multiple campuses and programs"},
			{Title: "Merojob Education Jobs", URL: "https://www.merojob.com", Description: "Find university teaching jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Choose your academic discipline and Excel in it", Description: "Pick a subject you are passionate about and could study for the rest of your life. Excel in your Bachelor's degree — aim for high grades, especially in your major. Get to know your professors. Ask them about their research and career paths. Read academic journals in your field. The foundation of an academic career is deep knowledge of your discipline. Your undergraduate years are when you build that foundation. A curious mind and love of learning are the most important qualities for a future professor.",
			Duration: "3-4 years", Links: []roadmapLink{
				{Title: "UGC Nepal - Higher education programs", URL: "https://www.ugc.gov.np"},
				{Title: "Choosing an academic discipline (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+choose+a+major+nepal"},
				{Title: "Google Scholar - Academic research search", URL: "https://scholar.google.com"},
			}},
			{StepNumber: 2, Title: "Complete a Master's degree with distinction", Description: "A Master's degree is the minimum requirement for university teaching. Pursue a Master's in your field at a Nepali university or abroad. Write a thesis or dissertation that contributes new knowledge. Publish your first academic paper based on your thesis. Network with academics in your field. Attend conferences. The Master's degree is where you transition from student to scholar. You learn not just to consume knowledge but to create it. A strong Master's thesis can be the foundation of your future academic career.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "TU Central Department of your subject", URL: "https://www.tu.edu.np"},
				{Title: "How to write a Master's thesis (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+write+a+thesis+nepal"},
				{Title: "Academic publishing guide for beginners", URL: "https://www.youtube.com/results?search_query=how+to+publish+academic+paper"},
			}},
			{StepNumber: 3, Title: "Earn a PhD and build your research portfolio", Description: "A PhD is essential for permanent professor positions. You can pursue a PhD in Nepal (limited programs) or abroad (through scholarships like Fulbright, MEXT, Commonwealth, or Erasmus Mundus). Your PhD involves original research that advances knowledge in your field. While doing your PhD, gain teaching experience as a teaching assistant. Publish papers in reputable journals. Present at international conferences. The PhD is a long journey (3-6 years) but it is the gateway to a academic career. It demonstrates that you can conduct independent research at the highest level.",
			Duration: "3-6 years", Links: []roadmapLink{
				{Title: "Fulbright Scholarship for Nepali students", URL: "https://www.fulbrightnepal.org.np"},
				{Title: "PhD opportunities in Nepal (UGC)", URL: "https://www.ugc.gov.np"},
				{Title: "How to get a PhD scholarship abroad (YouTube)", URL: "https://www.youtube.com/results?search_query=phd+scholarship+nepal+guide"},
			}},
			{StepNumber: 4, Title: "Apply for assistant professor positions", Description: "Apply for faculty positions at universities and colleges in Nepal. Submit your CV, research publications, and teaching philosophy. The hiring process may include a demonstration lecture and interview. Start as an Assistant Professor on a contract or tenure-track basis. Build your teaching skills. Develop new courses. Advise graduate students. Continue your research and publication. The first few years as a professor are busy but exciting. You are finally living your dream of being an academic.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Find university jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "How to apply for professor positions (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+become+a+professor+nepal"},
				{Title: "Academic CV and interview preparation", URL: "https://www.youtube.com/results?search_query=academic+interview+tips+nepal"},
			}},
			{StepNumber: 5, Title: "Get promoted through research and service", Description: "The academic promotion path is: Assistant Professor, Associate Professor, then Professor. Promotion requires a strong record of research publications, teaching excellence, student mentoring, and university service. Publish regularly in peer-reviewed journals. Apply for research grants. Supervise PhD students. Serve on academic committees. The journey from Assistant to full Professor takes 8-15 years of consistent work. Each promotion brings more respect, higher salary, and greater influence in your field.",
			Duration: "5-15 years", Links: []roadmapLink{
				{Title: "UGC Nepal - Faculty promotion criteria", URL: "https://www.ugc.gov.np"},
				{Title: "Research grant opportunities in Nepal", URL: "https://www.ugc.gov.np"},
				{Title: "Tips for academic promotion (YouTube)", URL: "https://www.youtube.com/results?search_query=academic+promotion+tips"},
			}},
			{StepNumber: 6, Title: "Become a leader in your field and institution", Description: "Senior professors can become department heads, deans, or even vice-chancellors of universities. They advise government on policy, lead major research projects, and represent Nepal in international academic forums. Some professors become public intellectuals, writing for newspapers and appearing on television. Others consult for international organizations like UNESCO, WHO, or the World Bank. A professor's influence extends far beyond the classroom. You are not just teaching students — you are building Nepal's future intellectual capacity. Your legacy lives on in every student you taught and every paper you published.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "University administration career path", URL: "https://www.ugc.gov.np"},
				{Title: "Academic leadership in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=academic+leadership+nepal"},
				{Title: "International research collaborations for Nepal scholars", URL: "https://www.youtube.com/results?search_query=international+research+nepal"},
			}},
		},
	}
}

func montessoriTeacher() careerSeed {
	return careerSeed{
		CategoryName: "Education",
		CategorySlug: "education",
		CategoryIcon: "🧸",
		Title:        "Montessori Teacher",
		Slug:         "montessori-teacher",
		Summary:      "Montessori teachers guide young children (ages 2-6) through self-directed learning activities using special materials and a child-centered approach.",
		Description:  "A Montessori teacher guides young children using the Montessori method developed by Dr. Maria Montessori. Instead of traditional lectures, Montessori teachers prepare an environment with special learning materials and allow children to choose their own activities. The teacher observes, guides, and introduces new materials when the child is ready. In Nepal, Montessori schools are growing in popularity, especially in Kathmandu and major cities. Parents value the Montessori approach for developing independence, concentration, and a love of learning. Montessori teachers need specialized training from a recognized Montessori training center. The Montessori Training Center of Nepal (MTCN) and other institutes offer certified programs. This career is ideal for people who love young children and believe in nurturing their natural curiosity.",
		DailyTasks: []string{
			"Prepare the classroom environment with Montessori materials",
			"Observe children and note their interests and progress",
			"Present Montessori materials to individuals or small groups",
			"Guide children in choosing their own work activities",
			"Maintain classroom order and model respectful behavior",
			"Communicate with parents about their child's development",
			"Care for the physical needs of young children (toilet, snacks, rest)",
		},
		Skills: []string{
			"Patience and gentleness with young children",
			"Observation skills to understand each child's needs",
			"Knowledge of Montessori philosophy and materials",
			"Classroom environment preparation",
			"Communication with parents and caregivers",
			"Creativity in making and using learning materials",
			"Calm and respectful approach to discipline",
		},
		SalaryMin:   200000,
		SalaryMax:   600000,
		Difficulty:  2,
		FutureProof: 68,
		EducationReq: "Montessori Teacher Training diploma (typically 1 year) from a recognized Montessori training center. SLC/SEE pass minimum. Early Childhood Development (ECD) training is also valuable. Some schools prefer candidates with +2 or Bachelor's degree.",
		Outlook:      "Montessori education is growing in Nepal as parents seek alternatives to traditional schooling. The number of Montessori schools is increasing, especially in Kathmandu Valley. Qualified Montessori teachers are in demand. The career offers deep satisfaction as you witness young children develop independence and confidence.",
		Tags:         []string{"education", "children", "montessori", "early-childhood", "teaching"},
		Resources: []resourceSeed{
			{Title: "Montessori Training Center of Nepal", URL: "https://www.montessori.edu.np", Description: "Official Montessori teacher training and certification in Nepal"},
			{Title: "Ministry of Education - ECD programs", URL: "https://www.moe.gov.np", Description: "Early Childhood Development programs and standards in Nepal"},
			{Title: "Merojob Education Jobs", URL: "https://www.merojob.com", Description: "Find Montessori teacher jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "See if you love working with young children", Description: "Spend time with children aged 2-6. Babysit, volunteer at a preschool, or help with children at your local temple or community center. Observe how young children learn and play. Ask yourself: Do I have endless patience? Can I stay calm when children are challenging? Do I enjoy seeing the world through a child's eyes? Working with young children is rewarding but requires tremendous patience and energy. If being around children makes you happy, this could be your path.",
			Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Montessori Training Center Nepal - Resources", URL: "https://www.montessori.edu.np"},
				{Title: "Working with young children (YouTube)", URL: "https://www.youtube.com/results?search_query=early+childhood+education+career"},
				{Title: "Volunteer at a preschool in Nepal", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 2, Title: "Complete Montessori Teacher Training", Description: "Enroll in a certified Montessori teacher training program. The Montessori Training Center of Nepal offers diploma courses (6-12 months). You will learn Montessori philosophy, child development, material presentations, classroom management, and observation techniques. Training includes both theory and supervised practice with children. This training changes how you see children — you learn to trust their natural desire to learn. The Montessori approach is not just a teaching method, it is a way of respecting children as capable individuals.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "MTCN - Montessori Diploma program", URL: "https://www.montessori.edu.np"},
				{Title: "Montessori philosophy overview (YouTube)", URL: "https://www.youtube.com/results?search_query=montessori+philosophy+explained"},
				{Title: "Montessori material demonstrations", URL: "https://www.youtube.com/results?search_query=montessori+materials+demonstration"},
			}},
			{StepNumber: 3, Title: "Find a job at a Montessori school", Description: "Apply to Montessori schools in Kathmandu, Pokhara, and other cities. Many schools are always looking for trained Montessori teachers. Start as an assistant teacher if needed. Observe the lead teacher carefully. Learn how they interact with children, how they handle challenging moments, how they prepare the environment. Build relationships with parents — they are your partners in each child's development. Your calm, respectful manner will earn their trust. The Montessori classroom is a prepared environment where children grow at their own pace. Your role is to guide, not to push.",
			Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Find Montessori teaching jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Montessori schools in Kathmandu", URL: "https://www.montessori.edu.np"},
				{Title: "Montessori assistant teacher tips (YouTube)", URL: "https://www.youtube.com/results?search_query=montessori+assistant+teacher+role"},
			}},
			{StepNumber: 4, Title: "Deepen your understanding of child development", Description: "Study child development beyond Montessori: Piaget (cognitive development), Vygotsky (social learning), and attachment theory. Take workshops on positive discipline, brain development, and sensory integration. The more you understand about how children develop, the better you can support them. Attend Montessori conferences and connect with other Montessori teachers. Sharing experiences with colleagues helps you grow. Every child is unique and your understanding of development helps you meet each child where they are.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Child development courses (free online)", URL: "https://www.youtube.com/results?search_query=child+development+for+teachers"},
				{Title: "Positive discipline for Montessori teachers", URL: "https://www.youtube.com/results?search_query=positive+discipline+montessori"},
				{Title: "Montessori conferences and workshops", URL: "https://www.montessori.edu.np"},
			}},
			{StepNumber: 5, Title: "Specialize or become a lead teacher", Description: "After 2-3 years, become a lead teacher with your own classroom. Some Montessori teachers specialize in Infant/Toddler (0-3), Primary (3-6), or Elementary (6-12) levels. Each level requires additional training. Lead teachers earn more and have more responsibility for designing the learning environment. Some Montessori teachers become trainers themselves, teaching new Montessori teachers at training centers. Share your knowledge and help grow the Montessori community in Nepal.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "Advanced Montessori training (3-6 level)", URL: "https://www.montessori.edu.np"},
				{Title: "Infant/toddler Montessori certification", URL: "https://www.montessori.edu.np"},
				{Title: "Becoming a Montessori teacher trainer", URL: "https://www.youtube.com/results?search_query=montessori+teacher+trainer+nepal"},
			}},
			{StepNumber: 6, Title: "Open your own Montessori school", Description: "Many experienced Montessori teachers eventually open their own school. Nepal's growing middle class is looking for quality early childhood education. Starting a Montessori school requires finding a suitable space, buying materials, hiring trained staff, and getting approval from the local government. It is a big step but very rewarding. You create a space where children can grow with dignity and joy. A Montessori school can start small — even in your own home with a few children. The Montessori approach changes children's lives and you can be the person who brings it to your community.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a Montessori school in Nepal", URL: "https://www.montessori.edu.np"},
				{Title: "Montessori school registration requirements", URL: "https://www.moe.gov.np"},
				{Title: "Montessori business guide (YouTube)", URL: "https://www.youtube.com/results?search_query=start+montessori+school+nepal"},
			}},
		},
	}
}

func englishInstructor() careerSeed {
	return careerSeed{
		CategoryName: "Education",
		CategorySlug: "education",
		CategoryIcon: "📝",
		Title:        "English Language Instructor",
		Slug:         "english-instructor",
		Summary:      "English instructors teach English to Nepali students and professionals. Good English skills open doors to better jobs, study abroad, and international opportunities.",
		Description:  "An English language instructor teaches English to students of all ages. In Nepal, English is a crucial skill for higher education, international business, tourism, and studying abroad. English instructors work in schools, language institutes (like British Council, ACE Institute, or English Access programs), universities, and as private tutors. Many Nepalis want to learn English for studying abroad (IELTS/TOEFL preparation), for jobs in tourism and hospitality, or for professional communication. English instructors can earn good income, especially those specialized in test preparation (IELTS, TOEFL, PTE, GRE). The demand for English teachers in Nepal is very high and stable.",
		DailyTasks: []string{
			"Plan and teach English lessons focusing on speaking, listening, reading, writing",
			"Assess students' English levels and track progress",
			"Prepare students for English proficiency tests (IELTS, TOEFL, PTE)",
			"Create engaging activities like role-plays, discussions, and games",
			"Grade assignments and provide feedback on language use",
			"Develop curriculum and select teaching materials",
			"Give one-on-one tutoring sessions for additional support",
		},
		Skills: []string{
			"Excellent English language proficiency (spoken and written)",
			"Teaching methodology for language acquisition",
			"Pronunciation and grammar knowledge",
			"Lesson planning and curriculum development",
			"Patience and encouragement for struggling learners",
			"Cross-cultural communication skills",
			"Test preparation strategies (IELTS, TOEFL)",
		},
		SalaryMin:   200000,
		SalaryMax:   1000000,
		Difficulty:  2,
		FutureProof: 70,
		EducationReq: "Bachelor's degree in English or Education preferred. TEFL/TESOL/CELTA certification highly valued. For IELTS preparation, IELTS Band 8+ or official IELTS trainer certification required. Native-level English fluency essential.",
		Outlook:      "English education is a massive industry in Nepal. Thousands of Nepalis take IELTS and TOEFL every year to study abroad. Schools at all levels need English teachers. Online English teaching to international students is also growing. Test preparation (IELTS, PTE) instructors can earn very high rates, especially those with proven results.",
		Tags:         []string{"education", "english", "teaching", "test-prep", "growing"},
		Resources: []resourceSeed{
			{Title: "British Council Nepal - Teaching Resources", URL: "https://www.britishcouncil.org.np", Description: "Free teaching resources and professional development for English teachers"},
			{Title: "TEFL/TESOL Certification Online", URL: "https://www.tefl.org", Description: "International English teaching certification accepted in Nepal"},
			{Title: "Merojob Education Jobs", URL: "https://www.merojob.com", Description: "Find English teacher jobs in Nepal"},
			{Title: "IELTS Official - Teaching Resources", URL: "https://www.ielts.org", Description: "Official IELTS preparation materials for teachers"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Master English yourself first", Description: "You cannot teach what you do not know well. Achieve excellent English yourself. Read English books, newspapers (Kathmandu Post, Himalayan Times), and websites. Watch English movies and TV shows. Practice speaking with fluent speakers. Aim for native-level fluency. Your own English skills are your most important teaching tool. If you plan to teach IELTS preparation, take the IELTS test yourself and aim for Band 8 or higher. Experience as a test-taker makes you a better test-prep teacher.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Kathmandu Post - English newspaper", URL: "https://kathmandupost.com"},
				{Title: "BBC Learning English (free)", URL: "https://www.bbc.co.uk/learningenglish"},
				{Title: "IELTS practice and preparation (British Council)", URL: "https://www.britishcouncil.org.np"},
			}},
			{StepNumber: 2, Title: "Get English teaching certification", Description: "Get a recognized English teaching certification: TEFL (Teaching English as a Foreign Language), TESOL (Teaching English to Speakers of Other Languages), or CELTA (Certificate in English Language Teaching to Adults). These are internationally recognized and accepted in Nepal. Courses can be taken online or in person. The British Council in Nepal offers CELTA, which is the gold standard. A certification not only teaches you methodology but also makes you employable. Language institutes prefer certified teachers.",
			Duration: "2-6 months", Links: []roadmapLink{
				{Title: "British Council - CELTA in Nepal", URL: "https://www.britishcouncil.org.np"},
				{Title: "Online TEFL certification (recognized)", URL: "https://www.tefl.org"},
				{Title: "TESOL certification options", URL: "https://www.tesol.org"},
			}},
			{StepNumber: 3, Title: "Start teaching and build your experience", Description: "Get a job at a language institute, school, or start with private tutoring. Many English teachers start by giving private lessons at home. Advertise your services in your neighborhood, social media, or through word of mouth. Start with general English or conversation classes. As you gain experience, you will develop your own teaching style. Learn what works with different students: children need games and songs, teenagers need engagement, adults need practical communication skills. Be patient and encouraging. Language learning takes time and every student progresses at their own pace.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Find English teaching jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "English teaching techniques (YouTube)", URL: "https://www.youtube.com/results?search_query=english+teaching+methods+for+beginners"},
				{Title: "Starting as a private English tutor in Nepal", URL: "https://www.youtube.com/results?search_query=private+english+tutor+nepal"},
			}},
			{StepNumber: 4, Title: "Specialize in high-demand areas", Description: "Specialize in areas where English teachers are most needed: IELTS/TOEFL/PTE test preparation, business English for professionals, English for tourism and hospitality, or academic English for university study. Test preparation teachers can charge premium rates, especially if they have a track record of high-scoring students. Business English teaching for corporate clients is well-paid. Specialization makes you more valuable and allows you to charge higher rates. Become the go-to person for your specialty.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "IELTS teacher training (IDP/British Council)", URL: "https://www.ielts.org"},
				{Title: "Business English teaching certification", URL: "https://www.britishcouncil.org.np"},
				{Title: "Teaching English for tourism (YouTube)", URL: "https://www.youtube.com/results?search_query=english+for+tourism+teachers"},
			}},
			{StepNumber: 5, Title: "Start teaching online to reach more students", Description: "Online English teaching opens up opportunities beyond Nepal. Teach students from China, Japan, Korea, or Latin America through platforms like Cambly, Preply, or iTalki. Many Nepali English teachers successfully teach online to international students. The pay is usually higher than local rates. All you need is a good internet connection, a quiet room, and your teaching skills. Online teaching also gives you flexibility and reach. You can teach from home while building a global student base.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Cambly - Teach English online", URL: "https://www.cambly.com"},
				{Title: "Preply - Online tutoring platform", URL: "https://www.preply.com"},
				{Title: "Tips for online English teachers from Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=online+english+teaching+nepal"},
			}},
			{StepNumber: 6, Title: "Become a teacher trainer or open your own institute", Description: "Experienced English instructors can train new teachers, become academic coordinators, or open their own language institute. There is huge demand for English language training in Nepal. Opening a small institute with a few classrooms and qualified teachers can be a successful business. Some senior teachers write textbooks or develop curriculum for schools. English is the language of global opportunity and in Nepal, good English teachers change lives. Every student who passes the IELTS because of you, every person who communicates confidently in English because of your teaching — that is your legacy.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a language institute in Nepal", URL: "https://www.moe.gov.np"},
				{Title: "Teacher training and mentoring (YouTube)", URL: "https://www.youtube.com/results?search_query=english+teacher+training+nepal"},
				{Title: "Curriculum development for English programs", URL: "https://www.britishcouncil.org.np"},
			}},
		},
	}
}

func onlineTutor() careerSeed {
	return careerSeed{
		CategoryName: "Education",
		CategorySlug: "education",
		CategoryIcon: "💻",
		Title:        "Online Tutor",
		Slug:         "online-tutor",
		Summary:      "Online tutors teach students over the internet using video calls. They can teach any subject — math, science, English, music, or test preparation.",
		Description:  "An online tutor teaches students over the internet using platforms like Zoom, Skype, or dedicated tutoring platforms. Online tutoring has grown enormously in Nepal, especially after the COVID-19 pandemic. Tutors can teach any subject: school subjects (math, science, English), test preparation (IELTS, SAT, GRE, Loksewa), computer skills, music, or languages. Online tutoring allows you to reach students anywhere in Nepal or even internationally. You can work from home and set your own schedule. The best part is that you do not need a fancy office or classroom — just a computer, good internet, and your knowledge. Many Nepali students seek online tutors because they can learn from the comfort of home.",
		DailyTasks: []string{
			"Prepare lesson materials and activities for online sessions",
			"Conduct live tutoring sessions via video calls",
			"Explain concepts clearly using screen sharing and digital tools",
			"Assign and grade homework and practice problems",
			"Track student progress and adjust teaching approach",
			"Respond to student questions between sessions",
			"Market your services on social media and tutoring platforms",
		},
		Skills: []string{
			"Expert knowledge in the subject you teach",
			"Clear communication and explanation skills",
			"Comfort with technology (video calls, screen sharing, digital whiteboards)",
			"Patience and the ability to adapt to each student's learning style",
			"Time management and scheduling",
			"Basic marketing and client management",
			"Reliable internet connection and technical setup",
		},
		SalaryMin:   150000,
		SalaryMax:   900000,
		Difficulty:  2,
		FutureProof: 75,
		EducationReq: "SLC/SEE pass minimum. Bachelor's degree or higher in the subject you teach preferred. Teaching or tutoring experience valuable. Any subject expertise can be turned into a tutoring business. Specific certifications for test prep (IELTS, SAT) helpful.",
		Outlook:      "Online tutoring in Nepal is growing rapidly. More families are willing to pay for personalized online education. International tutoring platforms connect Nepali tutors with students worldwide. Subjects like math, science, English, and test preparation have the highest demand. The market is competitive but good tutors with strong reputations can build a stable client base.",
		Tags:         []string{"education", "online", "freelance", "flexible", "remote-work"},
		Resources: []resourceSeed{
			{Title: "Preply - Become an Online Tutor", URL: "https://www.preply.com", Description: "Global platform connecting tutors with students"},
			{Title: "Chegg Tutors - Online Tutoring", URL: "https://www.chegg.com/tutors", Description: "Online tutoring platform for various subjects"},
			{Title: "Merojob Education Jobs", URL: "https://www.merojob.com", Description: "Find online tutoring jobs in Nepal"},
			{Title: "Google Classroom - Free Teaching Tools", URL: "https://classroom.google.com", Description: "Free platform to manage online classes and assignments"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Identify what you can teach well", Description: "Think about what subjects you know best. Are you good at math? Science? English? Music? Computer skills? Test preparation? Choose a subject you are confident teaching and enjoy explaining. Your enthusiasm for the subject will motivate your students. Consider what subjects have the most demand in Nepal: school-level math and science, English, IELTS/TOEFL preparation, computer programming, and Loksewa preparation. Pick something you know well and enjoy. Passion for your subject is contagious — students catch it.",
			Duration: "1-2 weeks", Links: []roadmapLink{
				{Title: "High-demand tutoring subjects in Nepal", URL: "https://www.merojob.com"},
				{Title: "How to choose your tutoring niche (YouTube)", URL: "https://www.youtube.com/results?search_query=choose+tutoring+subject+nepal"},
				{Title: "Online tutoring platforms comparison", URL: "https://www.preply.com"},
			}},
			{StepNumber: 2, Title: "Set up your home teaching space and get the right tools", Description: "You need a quiet room with good lighting, a computer or laptop, a reliable internet connection, a webcam, and a headset with microphone. Learn to use video calling tools (Zoom, Google Meet, Skype). Practice using screen sharing and a digital whiteboard. Prepare your teaching materials: sample problems, presentations, worksheets. Your setup does not need to be fancy but it needs to work reliably. Test everything before your first paid session. A session ruined by technical problems is a bad first impression.",
			Duration: "1-2 weeks", Links: []roadmapLink{
				{Title: "Setting up for online tutoring (YouTube)", URL: "https://www.youtube.com/results?search_query=online+tutor+setup+guide"},
				{Title: "Zoom for online teaching - Tutorial", URL: "https://www.youtube.com/results?search_query=zoom+for+teachers+tutorial"},
				{Title: "Digital whiteboard tools for tutoring", URL: "https://www.youtube.com/results?search_query=digital+whiteboard+for+tutoring"},
			}},
			{StepNumber: 3, Title: "Find your first students and build your reputation", Description: "Start with students you know — neighbors, family friends, relatives. Offer your first few sessions at a discount or even free in exchange for testimonials. Advertise on Facebook, in local community groups, and on your personal social media. Create a simple page listing your subjects, qualifications, and rates. Word of mouth is powerful in Nepal. Each satisfied student will bring you more. Be reliable, prepared, and encouraging. Your reputation is everything in this business. A tutor who is patient, knowledgeable, and punctual will never lack students.",
			Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Marketing your tutoring service (YouTube)", URL: "https://www.youtube.com/results?search_query=market+tutoring+business+nepal"},
				{Title: "Facebook for small business - Guide", URL: "https://www.youtube.com/results?search_query=facebook+page+for+tutor+nepal"},
				{Title: "Getting tutoring referrals and reviews", URL: "https://www.youtube.com/results?search_query=get+tutoring+referrals+nepal"},
			}},
			{StepNumber: 4, Title: "Join online tutoring platforms to reach more students", Description: "Sign up for global tutoring platforms like Preply, Cambly, Chegg Tutors, or TutorMe. These platforms connect you with students from around the world. Create a compelling profile with your qualifications, teaching approach, and a friendly video introduction. Start with competitive rates to get reviews, then raise your prices as your reputation grows. International platforms often pay better than local rates. You can also teach on Nepali platforms like Edusanjal or through direct Facebook groups.",
			Duration: "1-2 months", Links: []roadmapLink{
				{Title: "Preply - Create your tutor profile", URL: "https://www.preply.com"},
				{Title: "Cambly - Apply as a tutor", URL: "https://www.cambly.com"},
				{Title: "Tips for success on tutoring platforms (YouTube)", URL: "https://www.youtube.com/results?search_query=success+on+preply+tips+nepal"},
			}},
			{StepNumber: 5, Title: "Create structured courses and raise your rates", Description: "Instead of just helping with homework, create structured courses: 'Complete SEE Math Preparation,' 'IELTS Speaking Mastery,' 'Python Programming for Beginners.' Structured courses are more valuable and you can charge more. Develop a curriculum with clear learning objectives and outcomes. Record video lessons that students can watch between sessions. Package your offerings into different tiers. Professional tutors who treat their work as a business earn more than those who just take whatever comes.",
			Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Creating online course curriculum (YouTube)", URL: "https://www.youtube.com/results?search_query=create+online+course+curriculum"},
				{Title: "Pricing strategies for online tutors", URL: "https://www.youtube.com/results?search_query=tutor+pricing+nepal"},
				{Title: "Google Classroom for structured courses", URL: "https://classroom.google.com"},
			}},
			{StepNumber: 6, Title: "Scale your tutoring business", Description: "Once you are fully booked, raise your rates or hire other tutors to work under your brand. Create video courses that students can buy and watch anytime — this generates passive income. Build a YouTube channel with free lessons to attract new students. Some successful online tutors in Nepal earn more than traditional teachers. Build a brand around your teaching. The online education market in Nepal is growing fast and there is room for quality tutors who genuinely help students succeed. Your impact goes beyond the subjects you teach — you help students build confidence and achieve their dreams.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Scaling your tutoring business (YouTube)", URL: "https://www.youtube.com/results?search_query=scale+tutoring+business+nepal"},
				{Title: "Creating and selling online courses", URL: "https://www.youtube.com/results?search_query=create+online+course+nepal"},
				{Title: "YouTube channel for tutors - Guide", URL: "https://www.youtube.com/results?search_query=start+youtube+channel+for+tutor"},
			}},
		},
	}
}

func banker() careerSeed {
	return careerSeed{
		CategoryName: "Finance & Business",
		CategorySlug: "finance-business",
		CategoryDesc: "Careers in banking, finance, business management, and entrepreneurship that keep Nepal's economy running.",
		CategoryIcon: "🏦",
		Title:        "Banker",
		Slug:         "banker",
		Summary:      "Bankers manage customer accounts, process transactions, provide loans, and help people and businesses with their financial needs.",
		Description:  "A banker works in a bank helping customers manage their money. They open accounts, process deposits and withdrawals, provide loans, issue credit/debit cards, handle foreign exchange, and advise customers on financial products. Nepal has many banks including commercial banks (Nabil, Himalayan, Global IME, etc.), development banks, and microfinance institutions. The banking sector in Nepal is well-regulated by Nepal Rastra Bank (the central bank). Banking is a prestigious and stable career with good benefits. Bankers can work in branch banking, corporate banking, credit analysis, risk management, or treasury. Career progression from teller to branch manager to senior management is well-defined.",
		DailyTasks: []string{
			"Process customer deposits, withdrawals, and fund transfers",
			"Open new accounts and maintain customer records",
			"Evaluate and process loan applications",
			"Sell banking products like insurance and investment schemes",
			"Resolve customer complaints and issues",
			"Follow regulatory compliance and reporting requirements",
			"Attend training on new banking products and regulations",
		},
		Skills: []string{
			"Customer service and communication",
			"Knowledge of banking products and services",
			"Numeracy and attention to detail",
			"Sales and cross-selling skills",
			"Understanding of Nepal Rastra Bank regulations",
			"Basic computer skills and banking software",
			"Ethical conduct and confidentiality",
		},
		SalaryMin:   350000,
		SalaryMax:   2000000,
		Difficulty:  3,
		FutureProof: 72,
		EducationReq: "Bachelor's degree in Business, Finance, Accounting, or Economics from a recognized university. MBA or Master's in Finance preferred for advancement. Must pass Nepal Rastra Bank licensing for certain roles.",
		Outlook:      "The banking sector in Nepal is well-established with stable employment. Digital banking is changing the industry but bankers with customer service and advisory skills remain essential. Competition for jobs is high but banking offers good career progression. Rural branches need staff, offering opportunities for those willing to work outside Kathmandu.",
		Tags:         []string{"finance", "banking", "stable", "customer-service", "office-job"},
		Resources: []resourceSeed{
			{Title: "Nepal Rastra Bank", URL: "https://www.nrb.org.np", Description: "Central bank of Nepal — banking regulations and career info"},
			{Title: "Nepal Bankers Association", URL: "https://www.nepalbankers.com.np", Description: "Resources and networking for banking professionals in Nepal"},
			{Title: "Merojob Banking Jobs", URL: "https://www.merojob.com", Description: "Find banking jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Get a bachelor's degree in business or finance", Description: "Study business, finance, accounting, or economics at a recognized university. Focus on understanding financial statements, banking principles, and Nepal's financial system. Good grades and a strong understanding of basics will help you in bank recruitment exams. Join banking-related clubs or societies. Read financial news regularly — the Kathmandu Post economic section, and follow Nepal Rastra Bank policies. The more you understand about banking, the better you will perform in interviews.",
			Duration: "3-4 years", Links: []roadmapLink{
				{Title: "Nepal Rastra Bank - Banking career info", URL: "https://www.nrb.org.np"},
				{Title: "Kathmandu Post - Financial news", URL: "https://kathmandupost.com"},
				{Title: "Business/Finance programs in Nepal", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 2, Title: "Apply for entry-level positions or bank exams", Description: "Most Nepali banks recruit through written exams and interviews. Look for job announcements on bank websites and Merojob. Common entry-level positions: teller, customer service representative, or assistant. Pass the bank's written exam (covering math, English, general knowledge, and banking awareness) and interview. Prepare well for these exams — they are competitive. Some banks provide training for new hires. Your first role will involve lots of customer interaction and cash handling. Accuracy and speed are essential.",
			Duration: "3-12 months", Links: []roadmapLink{
				{Title: "Find banking jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Bank exam preparation tips (YouTube)", URL: "https://www.youtube.com/results?search_query=nepal+bank+exam+preparation"},
				{Title: "Bank teller training and skills", URL: "https://www.youtube.com/results?search_query=bank+teller+training+nepal"},
			}},
			{StepNumber: 3, Title: "Get trained in banking operations and regulations", Description: "Learn your bank's systems, products, and procedures. Understand Nepal Rastra Bank's rules and regulations. Get trained in anti-money laundering (AML) and know-your-customer (KYC) procedures. These are critical compliance areas. Build speed and accuracy in your daily work. Learn about different loan products — personal loans, home loans, business loans, agricultural loans. The more products you understand, the better you can serve customers and cross-sell. Banking is heavily regulated and compliance knowledge protects you and the bank.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Nepal Rastra Bank - Regulations and guidelines", URL: "https://www.nrb.org.np"},
				{Title: "AML/KYC training for bankers (YouTube)", URL: "https://www.youtube.com/results?search_query=aml+kyc+training+banking"},
				{Title: "Nepal Bankers Association - Training programs", URL: "https://www.nepalbankers.com.np"},
			}},
			{StepNumber: 4, Title: "Get promoted to officer or credit officer", Description: "After 2-3 years of experience, aim for promotion to officer level. Officer roles include: credit officer (evaluating loan applications), relationship manager (managing customer relationships), operations officer (managing branch operations), or compliance officer. Each role requires different skills. Credit officers need strong analytical skills to assess loan risk. Relationship managers need sales and communication skills. Choose the path that fits your strengths. Officer-level positions come with significantly higher pay and more responsibility.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Credit analysis training for bankers (YouTube)", URL: "https://www.youtube.com/results?search_query=credit+analysis+nepal+banking"},
				{Title: "Relationship management in banking", URL: "https://www.youtube.com/results?search_query=business+relationship+manager+skills"},
				{Title: "Banking career progression in Nepal", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 5, Title: "Get professional certifications", Description: "Certifications boost your career: CAIIB (Certified Associate of Indian Institute of Bankers — recognized in Nepal too), CFA (Chartered Financial Analyst), or risk management certifications. An MBA in Finance can accelerate promotion to management. Many banks support employees in pursuing higher education and certifications. Consider specializing: treasury, trade finance, risk management, or digital banking. Specialists are more valuable and earn more. The banking sector in Nepal offers many paths — find yours and pursue it with focus.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "MBA in Finance programs Nepal", URL: "https://www.tu.edu.np"},
				{Title: "CFA program for Nepali bankers", URL: "https://www.cfainstitute.org"},
				{Title: "Nepal Bankers Association - Professional development", URL: "https://www.nepalbankers.com.np"},
			}},
			{StepNumber: 6, Title: "Become a branch manager or senior specialist", Description: "With 8-10+ years of experience, become a branch manager, department head, or senior specialist. Branch managers are responsible for the entire performance of a branch — meeting targets, managing staff, ensuring compliance, and growing the business. Senior specialists in corporate banking, trade finance, or treasury earn very well. Some bankers move to Nepal Rastra Bank or regulatory roles. Banking in Nepal remains a prestigious career with good income, benefits, and respect. Your work helps individuals and businesses achieve their financial goals and contributes to Nepal's economic development.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Branch manager role and responsibilities", URL: "https://www.nepalbankers.com.np"},
				{Title: "Corporate banking career path (YouTube)", URL: "https://www.youtube.com/results?search_query=corporate+banking+career+nepal"},
				{Title: "Nepal Rastra Bank career opportunities", URL: "https://www.nrb.org.np"},
			}},
		},
	}
}

func microfinanceOfficer() careerSeed {
	return careerSeed{
		CategoryName: "Finance & Business",
		CategorySlug: "finance-business",
		CategoryIcon: "💵",
		Title:        "Microfinance Officer",
		Slug:         "microfinance-officer",
		Summary:      "Microfinance officers provide small loans and financial services to low-income families and small business owners in rural and urban Nepal.",
		Description:  "A microfinance officer provides small loans (typically 10,000-200,000 NPR) and other financial services to low-income individuals, especially women, who cannot access traditional banking. They work for microfinance institutions (MFIs) like Nirdhan, Chhimek, or private microfinance companies. Microfinance is a powerful tool for poverty reduction in Nepal. Officers visit clients in their homes or workplaces, assess their needs, disburse loans, collect payments, and provide financial literacy training. They also offer savings accounts, insurance, and remittance services. This career requires traveling a lot, especially in rural areas. It is a deeply rewarding job for those who want to help poor communities improve their economic condition.",
		DailyTasks: []string{
			"Visit clients in their homes or businesses to assess loan needs",
			"Process loan applications and verify client information",
			"Disburse loans and collect regular repayments",
			"Conduct group meetings with women's saving groups",
			"Provide financial literacy training to clients",
			"Track loan performance and follow up on overdue payments",
			"Maintain client records and report to the branch office",
		},
		Skills: []string{
			"Understanding of microfinance principles and operations",
			"Field work and community mobilization skills",
			"Basic accounting and financial record keeping",
			"Communication skills in Nepali and local languages",
			"Empathy and understanding of poor communities",
			"Negotiation and conflict resolution",
			"Ability to work independently in rural areas",
		},
		SalaryMin:   250000,
		SalaryMax:   700000,
		Difficulty:  3,
		FutureProof: 70,
		EducationReq: "SLC/SEE pass minimum, PCL/+2 or Bachelor's preferred. Training from microfinance institutions. Many MFIs provide on-the-job training. Experience in community work is valuable.",
		Outlook:      "Microfinance is a large sector in Nepal reaching millions of poor families. The government supports microfinance as a poverty reduction tool. The sector is regulated by Nepal Rastra Bank. Digital microfinance is growing but field officers are still essential. Career progression to branch manager or area manager is possible with experience.",
		Tags:         []string{"finance", "microfinance", "rural", "helping-people", "field-work"},
		Resources: []resourceSeed{
			{Title: "Nepal Rastra Bank - Microfinance Guidelines", URL: "https://www.nrb.org.np", Description: "Regulations and guidelines for microfinance institutions in Nepal"},
			{Title: "Nirdhan Microfinance - Career Info", URL: "https://www.nirdhan.com.np", Description: "One of Nepal's largest microfinance institutions"},
			{Title: "Merojob Banking Jobs", URL: "https://www.merojob.com", Description: "Find microfinance officer jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Understand poverty and microfinance in Nepal", Description: "Read about microfinance and how it helps poor people. Learn about Nepal's microfinance sector and major institutions. Talk to microfinance officers if you know any. Ask them about their daily work and challenges. Visit a local microfinance branch and observe. Microfinance is not just about lending money — it is about empowering people, especially women. If you come from a rural background, you already understand the challenges your clients will face. That understanding is invaluable.",
			Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Nepal Rastra Bank - Microfinance overview", URL: "https://www.nrb.org.np"},
				{Title: "What is microfinance? (YouTube)", URL: "https://www.youtube.com/results?search_query=microfinance+explained+nepal"},
				{Title: "Nirdhan Microfinance - About us", URL: "https://www.nirdhan.com.np"},
			}},
			{StepNumber: 2, Title: "Complete your education and apply for microfinance jobs", Description: "Complete at least SLC/SEE. A +2 or Bachelor's degree improves your chances. Apply for microfinance officer positions at MFIs. Many MFIs recruit locally — they prefer officers who speak the local language and understand the community. During the interview, emphasize your desire to help poor communities, your communication skills, and your willingness to travel. Entry-level microfinance officers receive 2-4 weeks of training before starting field work. Learn quickly and ask questions.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Find microfinance jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Microfinance training programs in Nepal", URL: "https://www.nrb.org.np"},
				{Title: "Interview tips for microfinance jobs (YouTube)", URL: "https://www.youtube.com/results?search_query=microfinance+job+interview+nepal"},
			}},
			{StepNumber: 3, Title: "Learn the field — visit clients and communities", Description: "As a new officer, accompany experienced officers on their client visits. Learn how to assess a client's business and repayment capacity. Learn to conduct group meetings with women's saving and credit groups. Build trust with clients — they need to see you as helpful, not threatening. Learn to collect repayments with patience and firmness. Keep accurate records of every transaction. Field work is tiring but rewarding. Every loan you disburse can change a family's life. A small loan for a sewing machine, a rickshaw, or a small shop can create sustainable income for years.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Client assessment in microfinance (YouTube)", URL: "https://www.youtube.com/results?search_query=microfinance+client+assessment"},
				{Title: "Women's saving groups best practices", URL: "https://www.nirdhan.com.np"},
				{Title: "Rural field work safety tips", URL: "https://www.youtube.com/results?search_query=field+work+safety+nepal"},
			}},
			{StepNumber: 4, Title: "Build expertise in financial literacy and client training", Description: "Learn to train clients in basic financial literacy: budgeting, saving, understanding interest rates, managing debt. Many MFI clients have never used formal financial services before. They need patient education. Develop simple training materials using local language and examples. Teach clients about the importance of saving alongside borrowing. Financial literacy is as important as the loans themselves. A client who understands money management is more likely to succeed and repay on time.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Financial literacy training for rural communities", URL: "https://www.youtube.com/results?search_query=financial+literacy+training+nepal"},
				{Title: "Microfinance client education resources", URL: "https://www.nrb.org.np"},
				{Title: "Group training techniques (YouTube)", URL: "https://www.youtube.com/results?search_query=community+training+techniques+nepal"},
			}},
			{StepNumber: 5, Title: "Get promoted to senior officer or branch manager", Description: "After 3-5 years of good performance, become a senior officer or branch manager. Manage a team of 5-10 officers. Oversee a portfolio of thousands of clients. Ensure loan quality and portfolio performance. Report to area or regional managers. Branch managers earn more and have more responsibility. Some officers move into training roles, teaching new recruits. Others move to the head office in operations, risk management, or product development. The microfinance sector offers clear career paths for dedicated performers.",
			Duration: "3-5 years", Links: []roadmapLink{
				{Title: "Microfinance branch management (YouTube)", URL: "https://www.youtube.com/results?search_query=microfinance+branch+management+nepal"},
				{Title: "Leadership in microfinance institutions", URL: "https://www.nirdhan.com.np"},
				{Title: "Microfinance portfolio management training", URL: "https://www.nrb.org.np"},
			}},
			{StepNumber: 6, Title: "Go into policy, consulting, or start your own MFI", Description: "Experienced microfinance professionals can work for Nepal Rastra Bank regulating MFIs, consult for international development organizations (World Bank, UNDP, USAID), or start their own microfinance institution. Starting an MFI requires significant capital and regulatory approval but can be very rewarding. Some professionals move into related fields: financial inclusion, digital finance, or social entrepreneurship. Microfinance has transformed millions of lives in Nepal. As a microfinance officer, you see the impact of your work every day — families lifting themselves out of poverty, children going to school, women gaining independence. That is the true reward.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Financial inclusion careers in Nepal", URL: "https://www.nrb.org.np"},
				{Title: "International development and microfinance", URL: "https://www.worldbank.org/en/country/nepal"},
				{Title: "Social entrepreneurship in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=social+entrepreneurship+nepal"},
			}},
		},
	}
}

func insuranceAgent() careerSeed {
	return careerSeed{
		CategoryName: "Finance & Business", CategorySlug: "finance-business", CategoryIcon: "🛡️",
		Title: "Insurance Agent", Slug: "insurance-agent",
		Summary: "Insurance agents sell insurance policies that protect people and businesses from financial losses due to accidents, illness, death, and property damage.",
		Description: "An insurance agent sells insurance policies to individuals and businesses. They help clients understand their insurance needs and find the right coverage — life insurance, health insurance, motor insurance, property insurance, or business insurance. In Nepal, the insurance sector is growing rapidly. The Insurance Board of Nepal regulates all insurance activities. Agents earn commissions on policies they sell. This career offers high earning potential for motivated people. Many agents in Nepal work part-time while building their client base. Successful agents build long-term relationships with clients who renew policies year after year. It is a sales career that requires persistence, good communication, and trustworthiness.",
		DailyTasks: []string{"Contact potential clients to explain insurance products", "Assess clients' insurance needs and recommend appropriate policies", "Prepare policy documents and process applications", "Follow up with clients for policy renewals", "Process insurance claims for clients", "Keep records of clients and policies sold", "Stay updated on new insurance products and regulations"},
		Skills: []string{"Sales and persuasion skills", "Communication and relationship building", "Knowledge of insurance products and regulations", "Self-motivation and discipline (commission-based income)", "Basic math for calculating premiums and commissions", "Customer service for claims processing", "Networking and lead generation"},
		SalaryMin: 150000, SalaryMax: 1000000, Difficulty: 2, FutureProof: 68,
		EducationReq: "SLC/SEE pass minimum. Must pass the Insurance Agent licensing exam conducted by the Insurance Board of Nepal. Training provided by insurance companies. Good communication skills and local network are valuable.",
		Outlook: "Insurance awareness is growing in Nepal. More people are buying life, health, and motor insurance. The government is promoting insurance penetration. The sector is growing 15-20% annually. Insurance agents are always needed, especially those who can explain products simply and build trust. Commission rates are regulated by the Insurance Board.",
		Tags: []string{"finance", "insurance", "sales", "commission", "flexible"},
		Resources: []resourceSeed{
			{Title: "Insurance Board of Nepal", URL: "https://www.ibn.gov.np", Description: "Regulatory body for insurance in Nepal — licensing and guidelines"},
			{Title: "Nepal Insurance Authority", URL: "https://www.nia.gov.np", Description: "Resources and information for insurance professionals"},
			{Title: "Merojob Insurance Jobs", URL: "https://www.merojob.com", Description: "Find insurance agent jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Get licensed as an insurance agent", Description: "The Insurance Board of Nepal requires all agents to pass a licensing exam. Study the provided materials on insurance products, regulations, and ethics. The exam is not very difficult — most people pass with some study. Once licensed, choose an insurance company to represent. Many companies provide initial training and support. Your license must be renewed periodically with continuing education. Being licensed makes you a legal professional and builds client trust.", Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Insurance Board of Nepal - Agent licensing", URL: "https://www.ibn.gov.np"},
				{Title: "Insurance agent exam preparation (YouTube)", URL: "https://www.youtube.com/results?search_query=insurance+agent+exam+nepal"},
				{Title: "Choosing an insurance company to represent", URL: "https://www.nia.gov.np"},
			}},
			{StepNumber: 2, Title: "Learn about insurance products inside out", Description: "Study all the products your company offers: life insurance (term, endowment, whole life), health insurance, motor insurance, property insurance, and micro-insurance. Understand each product's features, benefits, exclusions, and premium calculations. Know which product is best for which type of client. Clients trust agents who explain things clearly. If you do not understand a product yourself, you cannot sell it. Keep product brochures and comparison charts handy. Knowledge is your most powerful sales tool.", Duration: "1-2 months", Links: []roadmapLink{
				{Title: "Life insurance products explained (YouTube)", URL: "https://www.youtube.com/results?search_query=life+insurance+products+nepal"},
				{Title: "Health insurance in Nepal - Guide", URL: "https://www.ibn.gov.np"},
				{Title: "Motor insurance coverage explained", URL: "https://www.youtube.com/results?search_query=motor+insurance+nepal+guide"},
			}},
			{StepNumber: 3, Title: "Build your client network starting with family and friends", Description: "Your first clients will be people who know and trust you: family members, friends, neighbors, and acquaintances. Contact everyone in your personal network. Explain insurance in simple terms. Do not pressure — educate. Help them understand why insurance is important for their family's security. Offer the best policy for their needs, not the most expensive one. Happy clients will refer you to their friends and family. Your reputation is everything in this business. One satisfied client can bring you ten more through referrals.", Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Building a client base as a new agent (YouTube)", URL: "https://www.youtube.com/results?search_query=insurance+agent+client+network+nepal"},
				{Title: "Sales techniques for insurance agents", URL: "https://www.youtube.com/results?search_query=insurance+selling+techniques+nepal"},
				{Title: "Getting referrals from satisfied clients", URL: "https://www.youtube.com/results?search_query=get+insurance+referrals+nepal"},
			}},
			{StepNumber: 4, Title: "Provide excellent service for renewals and claims", Description: "Keep in touch with your clients after the sale. Remind them about renewal dates. Help them process claims when needed — this is when they need you most. A client who had a smooth claim experience will be your customer for life and tell everyone they know. Keep records of all policies and renewals. Use a simple spreadsheet or app to track your clients and their policy dates. Good service leads to renewals and referrals. Insurance is a relationship business, not a one-time sale.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Claims processing guide for agents (YouTube)", URL: "https://www.youtube.com/results?search_query=insurance+claim+process+nepal"},
				{Title: "Client relationship management for agents", URL: "https://www.youtube.com/results?search_query=insurance+client+management"},
				{Title: "Insurance renewal strategies", URL: "https://www.youtube.com/results?search_query=insurance+renewal+tips+nepal"},
			}},
			{StepNumber: 5, Title: "Specialize or become a team leader", Description: "After 2-3 years, consider specializing: health insurance specialist, corporate insurance, or agricultural insurance. You can also become a team leader, recruiting and training new agents. Team leaders earn override commissions on their team's sales. Some agents focus on high-value life insurance policies for wealthy clients. Others focus on micro-insurance for rural communities. Find your niche and become the expert. Specialized agents earn more and face less competition.", Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Insurance specialization options (YouTube)", URL: "https://www.youtube.com/results?search_query=insurance+specialization+nepal"},
				{Title: "Building an insurance sales team", URL: "https://www.youtube.com/results?search_query=recruit+insurance+agents+nepal"},
				{Title: "Corporate insurance sales guide", URL: "https://www.ibn.gov.np"},
			}},
			{StepNumber: 6, Title: "Advance to agency manager or broker", Description: "Top-performing agents can become agency managers, managing a network of agents under a company. Others become insurance brokers (independent agents who work with multiple companies). Some go into insurance consultancy or underwriting. The insurance industry in Nepal is expanding rapidly with new companies and products entering the market. As an insurance agent, you provide peace of mind to families and businesses. When a family receives a claim payment after a loss, you know you made a difference. That is the real reward of this career.", Duration: "ongoing", Links: []roadmapLink{
				{Title: "Agency manager role in insurance", URL: "https://www.ibn.gov.np"},
				{Title: "Insurance broker license requirements", URL: "https://www.ibn.gov.np"},
				{Title: "Career progression in insurance (YouTube)", URL: "https://www.youtube.com/results?search_query=insurance+career+growth+nepal"},
			}},
		},
	}
}

func realEstateAgent() careerSeed {
	return careerSeed{
		CategoryName: "Finance & Business", CategorySlug: "finance-business", CategoryIcon: "🏠",
		Title: "Real Estate Agent", Slug: "real-estate-agent",
		Summary: "Real estate agents help people buy, sell, and rent properties like land, houses, and apartments in Nepal.",
		Description: "A real estate agent helps clients buy, sell, or rent property. In Nepal, the real estate market is active in Kathmandu Valley, Pokhara, Chitwan, and other growing cities. Agents help sellers price their property, market it, show it to buyers, and negotiate the deal. They help buyers find the right property, arrange viewings, and guide them through the purchase process. The job requires knowledge of property values, legal procedures, and negotiation skills. Real estate can be very lucrative during market booms. Many agents work independently or for real estate companies. Good agents build a reputation that brings repeat business and referrals.",
		DailyTasks: []string{"Research property listings and market prices", "Show properties to potential buyers", "List new properties on websites and social media", "Negotiate prices and terms between buyers and sellers", "Prepare and process property documents", "Network with potential clients and other agents", "Advise clients on property values and market conditions"},
		Skills: []string{"Knowledge of Nepal's property market and prices", "Negotiation and persuasion skills", "Sales and client relationship management", "Understanding of property laws and registration", "Marketing and social media skills", "Communication in Nepali and English", "Patience with long sales cycles"},
		SalaryMin: 200000, SalaryMax: 1500000, Difficulty: 3, FutureProof: 62,
		EducationReq: "SLC/SEE pass minimum. Real estate training from Nepal Real Estate Association. Knowledge of property laws and registration procedures essential. Good local network very valuable.",
		Outlook: "The real estate market in Nepal fluctuates with the economy. Urban areas especially Kathmandu Valley have steady demand. Property values have generally risen over time. New housing projects and apartment complexes are creating opportunities. The profession is not highly regulated in Nepal, so competition varies.",
		Tags: []string{"business", "real-estate", "sales", "flexible", "commission"},
		Resources: []resourceSeed{
			{Title: "Nepal Real Estate Association", URL: "https://www.nreanepal.com", Description: "Professional body for real estate agents in Nepal"},
			{Title: "Land Revenue Department Nepal", URL: "https://www.dolr.gov.np", Description: "Government property registration and land records"},
			{Title: "Merojob Sales Jobs", URL: "https://www.merojob.com", Description: "Find real estate agent jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn the property market in your area", Description: "Study property prices in your target area. Walk around different neighborhoods. Talk to people about property values. Read property listings online. Understand what makes a property valuable: location, access to road, water, electricity, school district, proximity to market. Knowledge of your local market is your most important asset. If you know Kathmandu valley or your home town well, you already have a head start. The best agents know their area better than anyone.",
			Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Nepal Real Estate Association - Market info", URL: "https://www.nreanepal.com"},
				{Title: "Property listing websites in Nepal", URL: "https://www.youtube.com/results?search_query=real+estate+nepal+guide"},
				{Title: "Understanding land valuation in Nepal", URL: "https://www.dolr.gov.np"},
			}},
			{StepNumber: 2, Title: "Learn property law and registration procedures", Description: "Understanding property law is essential. Learn about land ownership types (raikar, guthi, etc.), property registration process, taxes (capital gains, registration fees), and required documents. Mistakes in property deals can be costly. Clients trust agents who understand the legal side. Study the Land Revenue Department's procedures. Learn about building codes and zoning regulations in your city. This knowledge protects your clients and makes you a trusted advisor.",
			Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Land Revenue Department - Registration guide", URL: "https://www.dolr.gov.np"},
				{Title: "Property tax and registration fees Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=property+registration+nepal+guide"},
				{Title: "Nepal building codes and zoning", URL: "https://www.youtube.com/results?search_query=building+code+nepal"},
			}},
			{StepNumber: 3, Title: "Start with properties in your own network", Description: "Tell everyone you know that you are now a real estate agent. Family, friends, neighbors, former classmates — they all know people who want to buy or sell property. Ask them to spread the word. Offer to help a family member sell or rent their property at a discount. Build a portfolio of properties you can show. Take good photos. Write clear descriptions. Post on Facebook Marketplace and local groups. Your first few deals will teach you more than any course.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Real estate marketing on Facebook in Nepal", URL: "https://www.youtube.com/results?search_query=real+estate+facebook+marketing+nepal"},
				{Title: "Building a real estate client list (YouTube)", URL: "https://www.youtube.com/results?search_query=real+estate+client+list+nepal"},
				{Title: "Property photography tips", URL: "https://www.youtube.com/results?search_query=real+estate+photography+tips"},
			}},
			{StepNumber: 4, Title: "Learn negotiation and deal-closing skills", Description: "Real estate deals involve negotiation. Learn to understand what both parties want and find a middle ground. Study negotiation techniques. Learn to handle objections. Understand how to structure deals: payment terms, advance amounts, possession dates. Good negotiators close more deals and get better prices for their clients. Practice your negotiation skills in low-stakes situations. Every successful deal builds your confidence. A closed deal means happy clients and a commission earned.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Real estate negotiation tips (YouTube)", URL: "https://www.youtube.com/results?search_query=real+estate+negotiation+skills"},
				{Title: "Closing real estate deals effectively", URL: "https://www.youtube.com/results?search_query=close+real+estate+deals+nepal"},
				{Title: "Real estate contract drafting guide", URL: "https://www.dolr.gov.np"},
			}},
			{StepNumber: 5, Title: "Build an online presence and get reviews", Description: "Create a professional Facebook page and Instagram for your real estate business. Post regularly — new listings, market updates, tips for buyers and sellers. Ask satisfied clients to write reviews and refer you. A strong online presence brings clients to you without cold calling. Consider running Facebook ads for your listings. Many buyers in Nepal search for property online. Make sure they find you. Real estate is increasingly digital and agents who embrace technology have an advantage.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Social media for real estate agents (YouTube)", URL: "https://www.youtube.com/results?search_query=social+media+real+estate+nepal"},
				{Title: "Facebook ads for property listings", URL: "https://www.youtube.com/results?search_query=facebook+ads+real+estate+nepal"},
				{Title: "Getting Google reviews for your business", URL: "https://www.youtube.com/results?search_query=get+google+reviews+real+estate"},
			}},
			{StepNumber: 6, Title: "Build a team or agency", Description: "Successful agents can build a team of agents working under them. Start your own real estate agency. Hire and train other agents. Focus on a specialty: luxury properties, commercial real estate, land development, or property management. Real estate can provide excellent income, especially in growing markets. As an agent, you help people find their dream home, make a good investment, or sell their biggest asset. It is a responsibility and a privilege.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a real estate agency in Nepal", URL: "https://www.nreanepal.com"},
				{Title: "Commercial real estate opportunities Nepal", URL: "https://www.youtube.com/results?search_query=commercial+real+estate+nepal"},
				{Title: "Property management as a business", URL: "https://www.youtube.com/results?search_query=property+management+nepal"},
			}},
		},
	}
}

func cooperativeManager() careerSeed {
	return careerSeed{
		CategoryName: "Finance & Business", CategorySlug: "finance-business", CategoryIcon: "🤝",
		Title: "Cooperative Manager", Slug: "cooperative-manager",
		Summary: "Cooperative managers run savings and credit cooperatives that help communities save money and access loans at fair rates.",
		Description: "A cooperative manager runs the daily operations of a savings and credit cooperative. Cooperatives are community-owned financial organizations that provide savings accounts, loans, and other services to members. Nepal has thousands of cooperatives regulated by the Department of Cooperatives. Managers oversee member enrollment, savings collection, loan disbursement, record keeping, and compliance with cooperative laws. Cooperatives are especially important in rural areas where banks have limited presence. The manager is responsible for the cooperative's financial health and must ensure loans are repaid and members' savings are safe. This career requires trustworthiness, financial skills, and community service orientation.",
		DailyTasks: []string{"Manage member savings accounts and loan records", "Evaluate and approve loan applications from members", "Ensure timely collection of loan repayments", "Prepare financial reports for the board of directors", "Conduct annual general meetings with members", "Ensure compliance with cooperative regulations", "Train staff and supervise daily operations"},
		Skills: []string{"Financial management and accounting", "Knowledge of cooperative laws and regulations", "Leadership and staff management", "Loan evaluation and risk assessment", "Community relations and communication", "Record keeping and reporting", "Integrity and ethical management"},
		SalaryMin: 300000, SalaryMax: 900000, Difficulty: 3, FutureProof: 70,
		EducationReq: "Bachelor's in Business, Finance, or Accounting preferred. Training from the Department of Cooperatives. Experience in accounting or financial services valuable. Knowledge of cooperative principles essential. Local language skills important.",
		Outlook: "Cooperatives play a major role in Nepal's financial sector, especially in rural areas. The government supports the cooperative movement. Well-managed cooperatives thrive but the sector faces challenges from mismanagement in some institutions. Qualified and honest cooperative managers are in demand. The sector offers stable employment for trustworthy professionals.",
		Tags: []string{"finance", "community", "management", "rural", "cooperative"},
		Resources: []resourceSeed{
			{Title: "Department of Cooperatives Nepal", URL: "https://www.deoc.gov.np", Description: "Government department regulating cooperatives in Nepal"},
			{Title: "Nepal Cooperative Association", URL: "https://www.cooperativenepal.com", Description: "Resources and training for cooperative professionals"},
			{Title: "Merojob Finance Jobs", URL: "https://www.merojob.com", Description: "Find cooperative manager jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn about cooperative principles and management", Description: "Study the cooperative model — member-owned, member-governed, member-benefiting. Learn about the Cooperative Act of Nepal and regulations. Understand how cooperatives differ from banks. Visit a well-run cooperative and talk to the manager. Ask about their daily work, challenges, and rewards. If you believe in community-based finance and helping people help themselves, this is the right career for you. The cooperative movement in Nepal has a proud history of empowering communities.",
			Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Department of Cooperatives - Training", URL: "https://www.deoc.gov.np"},
				{Title: "Cooperative management principles (YouTube)", URL: "https://www.youtube.com/results?search_query=cooperative+management+nepal"},
				{Title: "Cooperative Act of Nepal overview", URL: "https://www.deoc.gov.np"},
			}},
			{StepNumber: 2, Title: "Get education in accounting and financial management", Description: "Study accounting, financial management, and cooperative management. A Bachelor's in Business or Finance is very helpful. The Department of Cooperatives offers specific training programs for cooperative managers. Learn to prepare financial statements, manage cash flow, and conduct internal audits. Understand how to evaluate loan applications — the 5 Cs of credit (Character, Capacity, Capital, Collateral, Conditions). Good financial management keeps the cooperative healthy and protects members' savings.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Cooperative management courses Nepal", URL: "https://www.deoc.gov.np"},
				{Title: "Accounting for cooperatives (YouTube)", URL: "https://www.youtube.com/results?search_query=accounting+for+cooperatives+nepal"},
				{Title: "Financial management basics cooperatives", URL: "https://www.youtube.com/results?search_query=financial+management+cooperatives"},
			}},
			{StepNumber: 3, Title: "Start working in a cooperative in any role", Description: "Join a cooperative as an accountant, loan officer, or customer service representative. Learn the cooperative's systems, products, and member base. Understand the community the cooperative serves. Build relationships with members. Learn to handle cash, process transactions, and maintain records accurately. Show initiative and reliability. Good staff members are noticed and promoted. Your understanding of the community is valuable — you know who is trustworthy and who might default on a loan.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Find cooperative jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Loan officer training for cooperatives (YouTube)", URL: "https://www.youtube.com/results?search_query=loan+officer+training+nepal"},
				{Title: "Cooperative accounting procedures", URL: "https://www.deoc.gov.np"},
			}},
			{StepNumber: 4, Title: "Get promoted to manager or start a new cooperative", Description: "With experience, become a cooperative manager. If you see an underserved community, help them start a new cooperative. The Department of Cooperatives provides guidance for registering new cooperatives. A group of at least 7 members can form a cooperative. Starting a cooperative is a big responsibility but deeply rewarding. You create a financial institution owned by the community that serves the community. Your leadership can transform the economic life of a village.",
			Duration: "2-5 years", Links: []roadmapLink{
				{Title: "How to register a cooperative in Nepal", URL: "https://www.deoc.gov.np"},
				{Title: "Cooperative manager responsibilities (YouTube)", URL: "https://www.youtube.com/results?search_query=cooperative+manager+role+nepal"},
				{Title: "Nepal Cooperative Association - Resources", URL: "https://www.cooperativenepal.com"},
			}},
			{StepNumber: 5, Title: "Get certified as a cooperative auditor or trainer", Description: "Experienced managers can become cooperative auditors (inspecting other cooperatives), trainers (teaching new managers), or consultants. Get certified by the Department of Cooperatives as a cooperative auditor. Help struggling cooperatives improve their management. Train boards of directors on their responsibilities. This advanced role offers variety and higher income. The cooperative sector needs professionals who can help maintain standards and rebuild trust after cases of mismanagement have affected the sector's reputation.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Cooperative auditor certification Nepal", URL: "https://www.deoc.gov.np"},
				{Title: "Cooperative training and consultancy (YouTube)", URL: "https://www.youtube.com/results?search_query=cooperative+training+nepal"},
				{Title: "Cooperative governance best practices", URL: "https://www.cooperativenepal.com"},
			}},
			{StepNumber: 6, Title: "Lead multiple cooperatives or move to policy", Description: "Top cooperative professionals manage multiple cooperatives or move into policy roles with the Department of Cooperatives or the Nepal Cooperative Association. Some become advisors to international development organizations working on financial inclusion. The cooperative sector in Nepal employs thousands of people and serves millions of members. As a cooperative manager, you practice the principle of 'people helping people.' You are not just running a business — you are building community wealth and economic democracy.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Cooperative policy and regulation careers", URL: "https://www.deoc.gov.np"},
				{Title: "International cooperative development careers", URL: "https://www.ica.coop"},
				{Title: "Nepal Cooperative Association - Leadership", URL: "https://www.cooperativenepal.com"},
			}},
		},
	}
}

func smallBusinessOwner() careerSeed {
	return careerSeed{
		CategoryName: "Finance & Business", CategorySlug: "finance-business", CategoryIcon: "🏪",
		Title: "Small Business Owner", Slug: "small-business-owner",
		Summary: "Small business owners run their own shops, restaurants, services, or small manufacturing businesses. They manage everything from operations to finance to marketing.",
		Description: "A small business owner runs their own business, whether it is a retail shop, restaurant, repair service, small factory, or consultancy. In Nepal, small businesses are the backbone of the economy — from the corner momo shop to the hardware store in New Road. Business owners handle everything: sourcing products, managing staff, serving customers, keeping accounts, and planning for growth. It is a challenging but rewarding path. You are your own boss. Your income depends on your effort and decisions. Many of Nepal's largest businesses started as small family enterprises. Entrepreneurship is deeply valued in Nepali culture.",
		DailyTasks: []string{"Open the shop/business and prepare for customers", "Serve customers and handle sales transactions", "Manage inventory and order supplies", "Handle accounting, bills, and financial records", "Market your business through social media and word of mouth", "Manage employees and schedule shifts", "Solve customer complaints and operational problems"},
		Skills: []string{"Basic accounting and financial management", "Customer service and communication", "Inventory and supply chain management", "Marketing and sales skills", "Problem-solving and decision-making", "Staff management and leadership", "Resilience and adaptability"},
		SalaryMin: 200000, SalaryMax: 2000000, Difficulty: 3, FutureProof: 65,
		EducationReq: "No formal degree required. Business training from CTEVT or management institutes helpful. Many successful business owners learned through experience. Basic literacy and numeracy essential. Specific licenses required for certain businesses (restaurant, pharmacy, etc.).",
		Outlook: "Small businesses make up the majority of Nepal's economy. The government has programs to support small and medium enterprises (SMEs) through the Ministry of Industry. Digital tools are making it easier to run a small business. Competition is high but opportunities exist in underserved areas and niche markets. E-commerce is creating new possibilities for small businesses.",
		Tags: []string{"business", "entrepreneur", "self-employed", "flexible", "risk"},
		Resources: []resourceSeed{
			{Title: "Ministry of Industry Nepal - SME Programs", URL: "https://www.moind.gov.np", Description: "Government programs supporting small and medium enterprises"},
			{Title: "Federation of Nepalese Chambers of Commerce", URL: "https://www.fncci.org", Description: "Business resources and networking for Nepali entrepreneurs"},
			{Title: "Merojob Business Jobs", URL: "https://www.merojob.com", Description: "Find business opportunities and jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Identify a business idea that matches your skills", Description: "Think about what you are good at and what people in your area need. Are you good at cooking? Open a small restaurant or food stall. Good at fixing things? Open a repair shop. Good at making handicrafts? Sell them online. Research your market — who are your customers, what do they need, what are they willing to pay. Talk to people who already run similar businesses. Learn from their experience. The best business ideas solve a real problem for real people. Start with an idea that is small enough to test without risking everything.",
			Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Small business ideas for Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=small+business+ideas+nepal"},
				{Title: "Market research for small businesses", URL: "https://www.youtube.com/results?search_query=market+research+nepal"},
				{Title: "Ministry of Industry - SME support", URL: "https://www.moind.gov.np"},
			}},
			{StepNumber: 2, Title: "Write a simple business plan", Description: "Write down your business idea on paper or in a notebook. What is your product or service? Who are your customers? Where will you operate? How much money do you need to start? How much will you charge? How much do you expect to earn? A simple plan helps you think through the details. It also helps if you need a loan from a bank or cooperative. Your plan does not have to be fancy but it must be realistic. Many businesses fail because the owner did not plan properly. Do not skip this step.",
			Duration: "2-4 weeks", Links: []roadmapLink{
				{Title: "Business plan template for Nepali startups", URL: "https://www.moind.gov.np"},
				{Title: "How to write a business plan (YouTube)", URL: "https://www.youtube.com/results?search_query=write+business+plan+nepal"},
				{Title: "Small business loan options in Nepal", URL: "https://www.nrb.org.np"},
			}},
			{StepNumber: 3, Title: "Register your business and get necessary licenses", Description: "Register your business with the Company Registrar's Office (for Pvt. Ltd.) or the local ward office (for small sole proprietorship). Get a PAN (Permanent Account Number) from the Inland Revenue Department. Obtain any specific licenses required for your type of business (restaurant license, health permit, etc.). Registered businesses can open bank accounts, issue bills, and pay taxes properly. Being registered builds trust with customers and suppliers. The registration process in Nepal has become simpler with online systems.",
			Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Company registration in Nepal (OCR)", URL: "https://www.ocr.gov.np"},
				{Title: "Inland Revenue Department - PAN registration", URL: "https://www.ird.gov.np"},
				{Title: "Business licenses in Nepal guide (YouTube)", URL: "https://www.youtube.com/results?search_query=business+license+nepal"},
			}},
			{StepNumber: 4, Title: "Open for business and find your first customers", Description: "Launch your business. Tell everyone you know. Offer a promotion for first-time customers. Be present every day. Listen to customer feedback and improve. Your first customers are your most important teachers. Be patient — most businesses take months to build a steady client base. Focus on quality and service. A happy customer tells others. A unhappy customer tells everyone. Keep track of every rupee — income and expenses. Good record keeping helps you understand if you are making a profit.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Customer service tips for small business (YouTube)", URL: "https://www.youtube.com/results?search_query=small+business+customer+service+nepal"},
				{Title: "Simple bookkeeping for small businesses", URL: "https://www.youtube.com/results?search_query=bookkeeping+small+business+nepal"},
				{Title: "Marketing your new business on a budget", URL: "https://www.youtube.com/results?search_query=market+small+business+nepal"},
			}},
			{StepNumber: 5, Title: "Scale up — expand products, hire staff, open new locations", Description: "When your business is running smoothly and making consistent profit, think about growth. Add new products or services. Hire staff so you are not doing everything yourself. Open a second location. Sell online to reach more customers. Building a business that works without you is the goal of every entrepreneur. Growth should be gradual — expand only when you have stable cash flow. Many small businesses fail because they grew too fast. Patience and steady growth build lasting businesses.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Scaling a small business in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=scale+small+business+nepal"},
				{Title: "Hiring employees in Nepal - Guide", URL: "https://www.youtube.com/results?search_query=hire+employees+nepal+small+business"},
				{Title: "E-commerce for small businesses Nepal", URL: "https://www.sastodeal.com"},
			}},
			{StepNumber: 6, Title: "Build a brand and give back to your community", Description: "Build a brand that people recognize and trust. Create a logo, packaging, and consistent service quality. Use social media to tell your business story. Give back to your community — sponsor local events, hire local people, support local causes. Businesses that are part of the community thrive in Nepal. Nepali customers are loyal to businesses they trust. Your small business can grow into something that supports your whole family, employs others, and contributes to Nepal's economy. Every big business started as a small idea with a determined owner who refused to give up.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Branding for small businesses (YouTube)", URL: "https://www.youtube.com/results?search_query=branding+small+business+nepal"},
				{Title: "Corporate social responsibility for SMEs", URL: "https://www.fncci.org"},
				{Title: "FNCCI - Business networking resources", URL: "https://www.fncci.org"},
			}},
		},
	}
}

func stockBroker() careerSeed {
	return careerSeed{
		CategoryName: "Finance & Business", CategorySlug: "finance-business", CategoryIcon: "📈",
		Title: "Stock Broker / Investor", Slug: "stock-broker",
		Summary: "Stock brokers help people buy and sell shares on the Nepal Stock Exchange (NEPSE). They provide investment advice and manage trading accounts.",
		Description: "A stock broker helps clients buy and sell shares of publicly traded companies on the Nepal Stock Exchange (NEPSE). They execute trades, provide market information, advise clients on investment decisions, and manage portfolios. Brokers work for member brokerage firms of NEPSE. The Nepali stock market has grown significantly with more people investing. Brokers need to be licensed by the Securities Board of Nepal (SEBON). The job requires understanding of financial markets, companies, and economic trends. Brokers earn commissions on trades. Experienced brokers also provide wealth management and portfolio advisory services. The market has ups and downs but long-term trends have been positive.",
		DailyTasks: []string{"Execute buy and sell orders for clients", "Provide market information and investment advice", "Monitor stock prices and market news", "Maintain client accounts and records", "Process IPO and FPO applications for clients", "Research company financials and market trends", "Comply with SEBON regulations and reporting"},
		Skills: []string{"Knowledge of stock markets and NEPSE operations", "Financial analysis and company research", "Customer service and communication", "Risk assessment and investment advisory", "Computer skills for trading platforms", "Patience during market volatility", "Ethical conduct and regulatory compliance"},
		SalaryMin: 300000, SalaryMax: 3000000, Difficulty: 4, FutureProof: 65,
		EducationReq: "Bachelor's degree in Business, Finance, or Economics. Must pass SEBON broker licensing exam. Certified Securities Dealer license required. MBA or CFA preferred for senior roles.",
		Outlook: "NEPSE has seen growing participation from retail investors. The government is promoting capital market development. SEBON is regulating the market more strictly to protect investors. The market is volatile but long-term potential is positive. Competition among brokers is increasing. Digital trading platforms are changing the industry.",
		Tags: []string{"finance", "investing", "stock-market", "high-salary", "regulatory"},
		Resources: []resourceSeed{
			{Title: "Securities Board of Nepal (SEBON)", URL: "https://www.sebon.gov.np", Description: "Regulatory body for Nepal's securities market"},
			{Title: "Nepal Stock Exchange (NEPSE)", URL: "https://www.nepalstock.com", Description: "Official stock exchange of Nepal with market data"},
			{Title: "Merojob Finance Jobs", URL: "https://www.merojob.com", Description: "Find stock broker jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn about stock markets and NEPSE", Description: "Read about how stock markets work. Study NEPSE trading hours, settlement process, and listed companies. Open a DEMAT account and start investing small amounts in blue-chip companies. Personal investing experience is the best teacher. Learn to read stock charts, company financial reports, and market news. Follow the daily market summary in newspapers. Understanding the market from an investor's perspective is essential before advising others. You cannot guide clients if you have never invested yourself.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "NEPSE - Stock market education", URL: "https://www.nepalstock.com"},
				{Title: "SEBON - Investor education resources", URL: "https://www.sebon.gov.np"},
				{Title: "Stock market basics for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=stock+market+nepal+for+beginners"},
			}},
			{StepNumber: 2, Title: "Get a degree in finance and SEBON licensing", Description: "Complete a Bachelor's in Finance, Business, or Economics. Study financial markets, investment analysis, corporate finance, and securities law. Then pass the SEBON broker licensing exam — it covers securities law, market operations, accounting, and ethics. Preparation courses are available. A broker license is required by law. Licensed brokers are registered with SEBON and can legally execute trades for clients. Your license is your professional credential. Protecting it means following strict ethical standards.",
			Duration: "3-5 years", Links: []roadmapLink{
				{Title: "SEBON - Broker licensing requirements", URL: "https://www.sebon.gov.np"},
				{Title: "Finance programs in Nepal (TU, KU, PU)", URL: "https://www.tu.edu.np"},
				{Title: "SEBON license exam preparation (YouTube)", URL: "https://www.youtube.com/results?search_query=sebon+license+exam+nepal"},
			}},
			{StepNumber: 3, Title: "Start as a junior broker or dealer", Description: "Join a NEPSE member brokerage firm. Start as a junior broker, dealer, or market analyst. Learn the firm's client base, trading systems, and research methodology. Shadow senior brokers. Learn to execute trades quickly and accurately. Build relationships with clients. Learn about different types of clients — some want quick trades, others want long-term investment advice. Each needs a different approach. Your first year is about learning the practical side of the business. Every trade teaches you something about the market.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Find broker jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Day in the life of a stock broker (YouTube)", URL: "https://www.youtube.com/results?search_query=stock+broker+nepal+day+in+life"},
				{Title: "NEPSE member brokerage list", URL: "https://www.nepalstock.com"},
			}},
			{StepNumber: 4, Title: "Build a client base through trust and performance", Description: "Your clients trust you with their money. That trust must be earned every day. Provide honest advice — do not recommend stocks just to earn commissions. Help clients understand risk. Educate them about long-term investing. Build relationships that last through market ups and downs. Satisfied clients refer their friends and family. A broker with 50 loyal clients can earn a good living. Focus on service, not just sales. In the stock market, honesty and good advice are the best marketing.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Building client trust as a broker (YouTube)", URL: "https://www.youtube.com/results?search_query=stock+broker+client+relationships"},
				{Title: "Investment advisory best practices", URL: "https://www.sebon.gov.np"},
				{Title: "Understanding client risk profiles", URL: "https://www.youtube.com/results?search_query=client+risk+profile+investing"},
			}},
			{StepNumber: 5, Title: "Get advanced certifications for portfolio management", Description: "Earn advanced certifications: CFA (Chartered Financial Analyst), or SEBON's Investment Advisor license. These allow you to manage client portfolios (not just execute trades) and charge advisory fees. Portfolio management is higher-value work than simple trade execution. Study technical analysis (chart patterns) and fundamental analysis (company valuation). The best brokers combine both approaches. Advanced certifications and proven performance allow you to charge higher fees and attract wealthier clients.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "CFA program for Nepali professionals", URL: "https://www.cfainstitute.org"},
				{Title: "SEBON - Investment Advisor license", URL: "https://www.sebon.gov.np"},
				{Title: "Technical analysis course (YouTube)", URL: "https://www.youtube.com/results?search_query=technical+analysis+nepal"},
			}},
			{StepNumber: 6, Title: "Open your own brokerage firm", Description: "Experienced brokers can apply to SEBON to open their own brokerage firm. This requires significant capital (depends on SEBON regulations), office space, and staff. Opening your own firm is a major step but offers the highest earning potential. Some brokers move into merchant banking, asset management, or investment banking. Others become market commentators on TV and radio. The Nepali capital market is developing and there are many opportunities for knowledgeable professionals. Remember: the market rewards patience and punishes greed. Good advice to clients is also good advice for your own career.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "SEBON - Broker firm registration", URL: "https://www.sebon.gov.np"},
				{Title: "Merchant banking career in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=merchant+banking+nepal"},
				{Title: "Investment banking opportunities Nepal", URL: "https://www.youtube.com/results?search_query=investment+banking+nepal"},
			}},
		},
	}
}

func remittanceAgent() careerSeed {
	return careerSeed{
		CategoryName: "Finance & Business", CategorySlug: "finance-business", CategoryIcon: "💸",
		Title: "Remittance Services Agent", Slug: "remittance-agent",
		Summary: "Remittance agents help people send and receive money from family members working abroad. Remittance is a huge part of Nepal's economy.",
		Description: "A remittance agent provides money transfer services to people sending or receiving money from abroad. Nepal receives billions of dollars in remittances every year from Nepalis working overseas (Gulf countries, Malaysia, Korea, Japan, UK, US, Australia, etc.). Remittance agents work for money transfer companies like Western Union, MoneyGram, IME, Prabhu Money Transfer, SCT, and others. They process incoming remittances and outgoing transfers. Agents earn commissions on each transaction. The business requires a license from Nepal Rastra Bank. Remittance agents often combine their service with other businesses like travel agencies, convenience stores, or mobile top-up services. It is a stable business because remittance flows are consistent and growing.",
		DailyTasks: []string{"Process incoming remittance payments to recipients", "Send money abroad for customers", "Verify customer identities and maintain records", "Manage cash float and currency exchange", "Handle customer inquiries about transfer status", "Ensure compliance with anti-money laundering rules", "Promote additional services like mobile top-up and bill payments"},
		Skills: []string{"Cash handling and accuracy", "Customer service and communication", "Knowledge of money transfer procedures", "Basic computer skills for transaction systems", "Understanding of exchange rates and fees", "Compliance awareness (AML, KYC)", "Patience explaining services to customers"},
		SalaryMin: 200000, SalaryMax: 800000, Difficulty: 2, FutureProof: 68,
		EducationReq: "SLC/SEE pass minimum. Training provided by money transfer companies. Must register as a remittance agent with Nepal Rastra Bank. Good reputation in the community is essential.",
		Outlook: "Remittance inflows to Nepal are consistently high (over 8 billion USD annually). Many Nepali families depend on remittances for daily needs. The business is stable and growing. Digital remittance options are increasing but cash-based services remain important, especially in rural areas. Competition among agents exists but the total market is large.",
		Tags: []string{"finance", "remittance", "community-service", "cash-business", "stable"},
		Resources: []resourceSeed{
			{Title: "Nepal Rastra Bank - Remittance Guidelines", URL: "https://www.nrb.org.np", Description: "Regulations for remittance agents and money transfer businesses"},
			{Title: "IME Limited - Agent Network", URL: "https://www.ime.com.np", Description: "One of Nepal's largest remittance companies — agent information"},
			{Title: "Merojob Finance Jobs", URL: "https://www.merojob.com", Description: "Find remittance agent jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Choose a remittance company to represent", Description: "Research major remittance companies operating in Nepal: IME, Western Union, MoneyGram, Prabhu, SCT, and others. Compare their commission rates, terms, and support. Choose a well-known company with good agent support. Contact their agent network department. Most have specific requirements for agents: minimum transaction volume, security deposit, and compliance training. Established companies provide training, signage, and marketing support. Being associated with a trusted brand helps attract customers.",
			Duration: "1-3 months", Links: []roadmapLink{
				{Title: "IME Limited - Become an agent", URL: "https://www.ime.com.np"},
				{Title: "Nepal Rastra Bank - Remittance agent registration", URL: "https://www.nrb.org.np"},
				{Title: "Money transfer agent comparison Nepal", URL: "https://www.youtube.com/results?search_query=remittance+agent+nepal+guide"},
			}},
			{StepNumber: 2, Title: "Get registered with NRB and set up your location", Description: "Register as a remittance agent with Nepal Rastra Bank. The process includes application, verification, and meeting capital requirements. Set up your shop in a location with good foot traffic — near a bus stop, market, or busy road. Your shop needs a computer, printer, internet connection, cash safe, and secure counter. Proper signage and branding are important for customer trust. Make your shop feel safe and professional. People will be handling cash and need to trust you.",
			Duration: "2-4 months", Links: []roadmapLink{
				{Title: "NRB - Remittance agent registration process", URL: "https://www.nrb.org.np"},
				{Title: "Setting up a remittance shop (YouTube)", URL: "https://www.youtube.com/results?search_query=remittance+shop+setup+nepal"},
				{Title: "Security best practices for cash handling", URL: "https://www.youtube.com/results?search_query=cash+handling+safety+nepal"},
			}},
			{StepNumber: 3, Title: "Learn the systems and serve your first customers", Description: "Learn your company's money transfer system thoroughly. Practice sending and receiving test transactions. Understand exchange rates, fees, and limits. Know the documents required for different types of transfers. Serve your first customers carefully and accurately. Double-check everything. A mistake in a remittance can cause serious problems for a family waiting for money. Be patient — many customers may not be familiar with the process and need clear explanations. Build a reputation for being fast, accurate, and trustworthy.",
			Duration: "1-2 months", Links: []roadmapLink{
				{Title: "Money transfer system training (YouTube)", URL: "https://www.youtube.com/results?search_query=remittance+system+training"},
				{Title: "Customer service for remittance agents", URL: "https://www.youtube.com/results?search_query=remittance+agent+customer+service"},
				{Title: "AML compliance for money transfer agents", URL: "https://www.nrb.org.np"},
			}},
			{StepNumber: 4, Title: "Add additional services to increase income", Description: "Maximize your location by offering additional services: mobile recharge (Ncell, NTC), utility bill payments, bus ticket booking, travel ticketing, photocopy, printing, and passport photos. Many customers who come for remittance will use other services too. Each additional service increases your income without much extra cost. Become a one-stop shop for everyday services in your neighborhood. Cross-selling is the key to profitability for remittance agents. The more services you offer, the more valuable you become to the community.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Services to add to a remittance business (YouTube)", URL: "https://www.youtube.com/results?search_query=additional+services+remittance+nepal"},
				{Title: "Mobile recharge agent registration Nepal", URL: "https://www.ncell.com.np"},
				{Title: "Bus ticketing agent opportunities Nepal", URL: "https://www.youtube.com/results?search_query=bus+ticket+agent+nepal"},
			}},
			{StepNumber: 5, Title: "Represent multiple companies and grow transaction volume", Description: "Once established, represent multiple money transfer companies to serve more customers. Some companies are better for certain corridors (e.g., better rates for Malaysia vs. Qatar). Having multiple options makes your shop the preferred choice. Build relationships with local banks for better cash management. Increase your transaction volume to qualify for higher commission rates. Loyal customers come regularly — know them by name and greet them warmly. In the remittance business, personal relationships build loyalty.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Multi-company remittance agency tips (YouTube)", URL: "https://www.youtube.com/results?search_query=multi+remittance+agency+nepal"},
				{Title: "Remittance volume growth strategies", URL: "https://www.youtube.com/results?search_query=grow+remittance+business+nepal"},
				{Title: "Bank partnerships for remittance agents", URL: "https://www.nrb.org.np"},
			}},
			{StepNumber: 6, Title: "Expand to multiple locations or mobile service", Description: "Successful agents open additional locations in different neighborhoods or partner with other shops. Some bring remittance services to rural areas where access is limited. Mobile remittance services (using a motorcycle to deliver cash to villages) serve areas without agent shops. The remittance needs of Nepal will continue for generations. As an agent, you help families receive the money they depend on. Every remittance you process connects a hardworking Nepali abroad with their family at home. That connection matters more than the commission you earn.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Expanding remittance locations Nepal", URL: "https://www.ime.com.np"},
				{Title: "Rural remittance services (YouTube)", URL: "https://www.youtube.com/results?search_query=rural+remittance+nepal"},
				{Title: "Remittance agent scaling strategies", URL: "https://www.nrb.org.np"},
			}},
		},
	}
}

func civilEngineer() careerSeed {
	return careerSeed{
		CategoryName: "Engineering & Construction", CategorySlug: "engineering-construction", CategoryDesc: "Careers in designing, building, and maintaining Nepal's infrastructure — roads, bridges, buildings, and water systems.", CategoryIcon: "🏗️",
		Title: "Civil Engineer", Slug: "civil-engineer",
		Summary: "Civil engineers design and supervise the construction of roads, bridges, buildings, water supply systems, and other infrastructure projects in Nepal.",
		Description: "A civil engineer designs, plans, and oversees construction projects. They work on roads, bridges, buildings, airports, water supply systems, irrigation projects, and hydropower plants — all essential for Nepal's development. Civil engineers work for government agencies (Department of Roads, Department of Urban Development), private construction companies, consulting firms, and international development organizations. The Nepal Engineering Council (NEC) licenses all engineers. With Nepal investing heavily in infrastructure, civil engineers are in high demand. The job requires strong technical knowledge, project management skills, and the ability to work on-site in challenging terrain. Specializations include structural, transportation, geotechnical, water resources, and environmental engineering.",
		DailyTasks: []string{"Design structures using engineering software (AutoCAD, STAAD, ETABS)", "Visit construction sites to inspect progress and quality", "Prepare project plans, budgets, and schedules", "Review contractor work and ensure compliance with specifications", "Coordinate with architects, surveyors, and other engineers", "Prepare technical reports and documentation", "Ensure safety standards are followed on construction sites"},
		Skills: []string{"Structural analysis and design", "Construction project management", "Knowledge of Nepal building codes and standards", "AutoCAD, STAAD, and other engineering software", "Surveying and site inspection", "Contract management and procurement", "Problem-solving and decision-making under pressure"},
		SalaryMin: 400000, SalaryMax: 2500000, Difficulty: 4, FutureProof: 80,
		EducationReq: "Bachelor's in Civil Engineering (BE/B.Tech) from recognized university (TU, KU, PU, or equivalent). Must register with Nepal Engineering Council. Master's degree for specialized roles. Nepal Engineering Council licensing exam required.",
		Outlook: "Nepal's infrastructure needs are enormous — roads, hydropower, airports, buildings, water systems. The government and international donors are investing heavily. Civil engineers are in high demand. The field offers stable employment and opportunities for career growth. Government jobs (through Loksewa) offer security. Private sector and consulting offer higher pay.",
		Tags: []string{"engineering", "construction", "infrastructure", "stable", "high-demand"},
		Resources: []resourceSeed{
			{Title: "Nepal Engineering Council", URL: "https://www.nec.gov.np", Description: "Licensing and regulatory body for engineers in Nepal"},
			{Title: "Department of Roads Nepal", URL: "https://www.dor.gov.np", Description: "Government department managing road infrastructure projects"},
			{Title: "Merojob Engineering Jobs", URL: "https://www.merojob.com", Description: "Find civil engineer jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Excel in math and physics in high school", Description: "Civil engineering requires strong math (calculus, trigonometry, algebra) and physics (mechanics, forces, materials). Focus on these subjects in SEE and +2 science. Read about famous structures and how they were built. Visit construction sites and observe. Your interest in how things are built is the foundation of this career. Good grades in science are essential for admission to engineering colleges. Every great structure started with someone who understood the math and physics behind it.",
			Duration: "2 years", Links: []roadmapLink{
				{Title: "Nepal Engineering Council - Approved colleges", URL: "https://www.nec.gov.np"},
				{Title: "Civil engineering overview (YouTube)", URL: "https://www.youtube.com/results?search_query=civil+engineering+career+nepal"},
				{Title: "Engineering entrance preparation Nepal", URL: "https://www.youtube.com/results?search_query=engineering+entrance+nepal+preparation"},
			}},
			{StepNumber: 2, Title: "Complete a Bachelor's in Civil Engineering", Description: "Enroll in a BE/B.Tech in Civil Engineering at a recognized college. Study structural engineering, geotechnical engineering, transportation engineering, water resources, and environmental engineering. Gain proficiency in design software. Complete internships with construction companies or engineering consultancies. Your final year project should solve a real Nepal problem — earthquake-resistant housing, rural road design, or water supply systems. Practical exposure during college is as important as classroom learning.",
			Duration: "4 years", Links: []roadmapLink{
				{Title: "TU Institute of Engineering - Civil programs", URL: "https://www.ioe.edu.np"},
				{Title: "Kathmandu University - Civil Engineering", URL: "https://www.ku.edu.np"},
				{Title: "AutoCAD tutorial for civil engineers (YouTube)", URL: "https://www.youtube.com/results?search_query=autocad+for+civil+engineers"},
			}},
			{StepNumber: 3, Title: "Get licensed with Nepal Engineering Council", Description: "After graduation, register with the Nepal Engineering Council (NEC). Submit your degree certificate, transcripts, and internship documents. Pass the NEC licensing exam. The license is mandatory to practice as an engineer in Nepal. It is your professional credential and must be renewed. Licensed engineers can sign off on projects, supervise construction, and take legal responsibility for their designs. Your NEC registration proves you meet Nepal's professional standards.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Nepal Engineering Council - Registration", URL: "https://www.nec.gov.np"},
				{Title: "NEC license exam preparation (YouTube)", URL: "https://www.youtube.com/results?search_query=nec+exam+nepal+preparation"},
				{Title: "Engineering Council code of ethics", URL: "https://www.nec.gov.np"},
			}},
			{StepNumber: 4, Title: "Start working and gaining field experience", Description: "Join a construction company, engineering consultancy, or government department as a junior engineer. Work under senior engineers. Learn project management, contract administration, and site supervision. Understand how designs are implemented in the field — reality often differs from drawings. Build your practical knowledge. Work on at least 2-3 complete projects (from design to completion) before considering yourself experienced. Keep a journal of lessons learned. Every project teaches something new. Be willing to work on-site, sometimes in remote areas.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Find civil engineer jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Construction site management tips (YouTube)", URL: "https://www.youtube.com/results?search_query=construction+site+management+nepal"},
				{Title: "Nepal standard specifications for construction", URL: "https://www.dor.gov.np"},
			}},
			{StepNumber: 5, Title: "Specialize and pursue a Master's degree", Description: "After 2-4 years, specialize: structural engineering (buildings and bridges), transportation engineering (roads and highways), water resources (hydropower, irrigation), geotechnical engineering (soil and foundations), or construction management. A Master's degree (ME/M.Sc.) in your specialization opens doors to senior positions. Specialized engineers are more valuable and earn more. Nepal's hydropower sector alone needs thousands of civil engineers. Find a specialization that matches Nepal's development needs and your interests.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "ME/M.Sc. civil engineering programs Nepal", URL: "https://www.ioe.edu.np"},
				{Title: "Hydropower engineering career Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=hydropower+engineering+nepal+career"},
				{Title: "Structural engineering specialization guide", URL: "https://www.youtube.com/results?search_query=structural+engineering+specialization"},
			}},
			{StepNumber: 6, Title: "Become a project manager or start your own firm", Description: "Experienced engineers become project managers overseeing multi-million rupee projects. Some become department heads in government (through Loksewa promotion) or partners in consulting firms. Others start their own engineering consultancy. Nepal needs qualified engineers to rebuild after earthquakes, build new infrastructure, and develop hydropower. The opportunities are vast. Your work as a civil engineer literally shapes Nepal's landscape. Roads connect villages, bridges cross rivers, buildings shelter families, and water systems bring health. You build the nation.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting an engineering consultancy in Nepal", URL: "https://www.nec.gov.np"},
				{Title: "Project management for civil engineers (YouTube)", URL: "https://www.youtube.com/results?search_query=project+management+civil+engineers+nepal"},
				{Title: "Loksewa civil engineer career path", URL: "https://www.psc.gov.np"},
			}},
		},
	}
}

func electricalEngineer() careerSeed {
	return careerSeed{
		CategoryName: "Engineering & Construction", CategorySlug: "engineering-construction", CategoryIcon: "⚡",
		Title: "Electrical Engineer", Slug: "electrical-engineer",
		Summary: "Electrical engineers design and maintain electrical systems — power generation, transmission, distribution, and electrical equipment for buildings and industry.",
		Description: "An electrical engineer works with electrical systems — from power generation (hydropower plants) to transmission lines, distribution networks, building wiring, and industrial electrical equipment. Nepal's massive hydropower development program creates huge demand for electrical engineers. They work for the Nepal Electricity Authority (NEA), hydropower developers, consulting firms, construction companies, and manufacturing industries. The job involves designing electrical systems, supervising installation, testing equipment, troubleshooting faults, and ensuring safety. Electrical engineers also work on renewable energy (solar), substations, and electrical equipment manufacturing. The Nepal Engineering Council licenses electrical engineers. The field offers excellent career prospects as Nepal expands its electricity infrastructure.",
		DailyTasks: []string{"Design electrical systems for buildings and industrial projects", "Supervise installation of electrical equipment and wiring", "Test and commission electrical systems and equipment", "Troubleshoot electrical faults and failures", "Prepare technical specifications and cost estimates", "Inspect electrical work for compliance with standards", "Coordinate with mechanical and civil engineers on projects"},
		Skills: []string{"Power system design and analysis", "Knowledge of electrical codes and safety standards", "AutoCAD electrical and design software", "Understanding of transformers, switchgear, and protection systems", "Project management and supervision", "Troubleshooting and problem-solving", "Knowledge of hydropower and renewable energy systems"},
		SalaryMin: 400000, SalaryMax: 2400000, Difficulty: 4, FutureProof: 82,
		EducationReq: "Bachelor's in Electrical Engineering (BE/B.Tech) from recognized university. Nepal Engineering Council registration required. Master's in Power Systems or Renewable Energy for specialized roles. NEA and hydropower sector specific training valuable.",
		Outlook: "Nepal's hydropower sector is booming with dozens of projects under construction. The government aims to generate 15,000 MW in the next decade. The Nepal Electricity Authority is expanding transmission and distribution networks. Solar energy is growing. Electrical engineers are in very high demand. The field offers excellent salaries and international opportunities.",
		Tags: []string{"engineering", "electrical", "hydropower", "high-demand", "energy"},
		Resources: []resourceSeed{
			{Title: "Nepal Engineering Council", URL: "https://www.nec.gov.np", Description: "Licensing for electrical engineers in Nepal"},
			{Title: "Nepal Electricity Authority", URL: "https://www.nea.org.np", Description: "National power utility — career and project information"},
			{Title: "Merojob Engineering Jobs", URL: "https://www.merojob.com", Description: "Find electrical engineer jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Build strong foundations in physics and math", Description: "Physics (especially electricity, magnetism, and circuits) and mathematics (calculus and algebra) are essential for electrical engineering. Focus on these in SEE and +2 science. Read about how electricity is generated and distributed. Visit a nearby hydropower plant or substation if possible. Understanding the basics of electricity from an early stage gives you a head start. Electrical engineering is at the heart of Nepal's development — from lighting homes to powering industries.",
			Duration: "2 years", Links: []roadmapLink{
				{Title: "Nepal Engineering Council - Approved colleges", URL: "https://www.nec.gov.np"},
				{Title: "Electrical engineering career overview (YouTube)", URL: "https://www.youtube.com/results?search_query=electrical+engineering+nepal+career"},
				{Title: "Engineering entrance exam preparation Nepal", URL: "https://www.youtube.com/results?search_query=engineering+entrance+exam+nepal"},
			}},
			{StepNumber: 2, Title: "Complete your Bachelor's in Electrical Engineering", Description: "Enroll in a BE program at a recognized college. Study power systems, electrical machines, control systems, electronics, and instrumentation. Gain hands-on experience with lab equipment. Learn design software like AutoCAD Electrical, ETAP, or DigSILENT. Complete internships with NEA, hydropower projects, or electrical manufacturing companies. Your final year project could focus on Nepal-specific challenges like rural electrification, solar mini-grids, or hydropower optimization. Practical experience combined with theory makes you job-ready.",
			Duration: "4 years", Links: []roadmapLink{
				{Title: "TU Institute of Engineering - Electrical programs", URL: "https://www.ioe.edu.np"},
				{Title: "Pokhara University - Electrical Engineering", URL: "https://www.pu.edu.np"},
				{Title: "Power system analysis basics (YouTube)", URL: "https://www.youtube.com/results?search_query=power+system+analysis+for+beginners"},
			}},
			{StepNumber: 3, Title: "Get NEC license and start your career", Description: "Register with the Nepal Engineering Council. Pass the licensing exam. Then apply for positions in the Nepal Electricity Authority, hydropower companies, consulting firms, or construction companies. Entry-level electrical engineers often start in project sites — supervising electrical installation, testing equipment, or assisting senior engineers. Be willing to work at remote project sites. Hydropower projects are often in rural areas with challenging living conditions. The experience you gain in the first 2-3 years is invaluable. Learn everything you can from senior engineers.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "NEC - Electrical engineer registration", URL: "https://www.nec.gov.np"},
				{Title: "Nepal Electricity Authority - Careers", URL: "https://www.nea.org.np"},
				{Title: "Find electrical engineering jobs (Merojob)", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 4, Title: "Get specialized training in hydropower or power systems", Description: "Nepal's hydropower boom creates demand for specialized skills. Take training in hydropower plant design, transmission line engineering, substation design, or protection system coordination. Learn about solar PV system design for the growing solar sector. Get certifications in relevant software and standards. Specialized electrical engineers earn significantly more. The Nepal Electricity Authority and independent power producers need engineers who understand the specific challenges of Nepal's power systems — from Himalayan rivers to urban distribution networks.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Hydropower engineering training Nepal", URL: "https://www.nea.org.np"},
				{Title: "Solar PV design course (YouTube)", URL: "https://www.youtube.com/results?search_query=solar+PV+design+training+nepal"},
				{Title: "Transmission line engineering guide", URL: "https://www.youtube.com/results?search_query=transmission+line+design+nepal"},
			}},
			{StepNumber: 5, Title: "Pursue a Master's in Power Systems or Renewable Energy", Description: "A Master's degree qualifies you for senior roles: senior engineer, project manager, department head. Specialize in power systems, renewable energy, or high voltage engineering. Master's graduates are preferred for design and planning roles at NEA, large hydropower projects, and international consulting firms. Some engineers pursue an MBA to move into management. Combining technical expertise with management skills makes you a strong candidate for leadership positions in Nepal's growing energy sector.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "ME Power Systems programs Nepal", URL: "https://www.ioe.edu.np"},
				{Title: "Renewable energy specialization (YouTube)", URL: "https://www.youtube.com/results?search_query=renewable+energy+engineering+nepal"},
				{Title: "MBA for engineers in Nepal", URL: "https://www.tu.edu.np"},
			}},
			{StepNumber: 6, Title: "Lead large projects or become a consultant", Description: "Experienced electrical engineers become project managers for hydropower projects worth billions of rupees. Some start their own electrical engineering consultancy. Others work as independent experts for international development banks (World Bank, ADB) funding Nepal's energy projects. The potential for career growth in Nepal's electrical engineering field is enormous — the country has one of the highest hydropower potentials in the world and is only scratching the surface. Your work as an electrical engineer brings light to homes, powers industries, and drives Nepal's economic growth.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Consulting career for electrical engineers (YouTube)", URL: "https://www.youtube.com/results?search_query=engineering+consultant+nepal"},
				{Title: "Hydropower project management", URL: "https://www.nea.org.np"},
				{Title: "International energy projects for Nepal engineers", URL: "https://www.youtube.com/results?search_query=international+hydropower+careers+nepal"},
			}},
		},
	}
}

func architect() careerSeed {
	return careerSeed{
		CategoryName: "Engineering & Construction", CategorySlug: "engineering-construction", CategoryIcon: "🏛️",
		Title: "Architect", Slug: "architect",
		Summary: "Architects design buildings and other structures. They create plans that are functional, safe, beautiful, and culturally appropriate for Nepal.",
		Description: "An architect plans and designs buildings and other structures. They create designs that are functional, safe, aesthetically pleasing, and suited to the local context. In Nepal, architects design homes, office buildings, schools, hospitals, hotels, and cultural centers. They must consider Nepal's earthquake risk, local materials, traditional architecture, and climate. Architects work with engineers, contractors, and clients throughout the construction process. The Nepal Engineering Council licenses architects. The profession combines art and science — creativity with technical knowledge. Nepal's reconstruction after the 2015 earthquake created many opportunities for architects. Traditional Nepali architecture (pagoda style, Newari courtyards) is world-famous and architects who understand it are especially valued.",
		DailyTasks: []string{"Meet clients to understand their needs and budget", "Create design concepts and sketches", "Develop detailed construction drawings using CAD software", "Prepare 3D models and visualizations", "Visit construction sites to ensure work matches design", "Coordinate with structural, electrical, and plumbing engineers", "Apply for building permits and ensure code compliance"},
		Skills: []string{"Architectural design and space planning", "CAD software (AutoCAD, Revit, SketchUp)", "Knowledge of Nepal building codes and earthquake-resistant design", "Creativity and visual imagination", "Communication with clients and contractors", "Understanding of traditional and modern architecture", "Project management and construction supervision"},
		SalaryMin: 400000, SalaryMax: 2000000, Difficulty: 4, FutureProof: 70,
		EducationReq: "Bachelor's in Architecture (B.Arch) from recognized university (TU, KU, PU). 5-year program including design studio. Must register with Nepal Engineering Council. Internship under registered architect required. Portfolio of design work essential for career.",
		Outlook: "Nepal's urban areas are growing rapidly, creating demand for architects. The post-earthquake reconstruction created awareness about safe building design. Heritage conservation is a growing field. Tourism infrastructure (hotels, resorts) needs architectural design. The profession is competitive but talented architects with good design sense and technical skills are always in demand.",
		Tags: []string{"architecture", "design", "creative", "construction", "engineering"},
		Resources: []resourceSeed{
			{Title: "Nepal Engineering Council - Architect Registration", URL: "https://www.nec.gov.np", Description: "Licensing and registration for architects in Nepal"},
			{Title: "Society of Nepalese Architects (SONA)", URL: "https://www.sonanepal.com", Description: "Professional body for architects in Nepal"},
			{Title: "Merojob Architecture Jobs", URL: "https://www.merojob.com", Description: "Find architect jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Develop your drawing and creative skills", Description: "Practice drawing and sketching buildings, spaces, and objects. Learn to observe your surroundings — how buildings relate to streets, how light enters rooms, how people use spaces. Study Nepal's traditional architecture: pagoda temples, Newari houses, Sherpa stone houses. These are world-class architectural heritage. Read about famous architects. Visit buildings under construction and talk to workers. Architecture is both art and science — develop both sides. Your portfolio of drawings and observations will help you get into architecture school.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Society of Nepalese Architects - Resources", URL: "https://www.sonanepal.com"},
				{Title: "Architecture sketching for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=architectural+sketching+for+beginners"},
				{Title: "Nepali architecture history and heritage", URL: "https://www.youtube.com/results?search_query=nepali+traditional+architecture"},
			}},
			{StepNumber: 2, Title: "Complete a Bachelor's in Architecture", Description: "B.Arch is a 5-year program. You will study design studio (core subject), architectural history, building technology, structures, environmental systems, and urban design. Design studio is where you spend most of your time — developing projects from concept to detailed design. Be prepared for long hours working on projects. Develop your own design style while learning from masters. Build a portfolio of your best work from each year. Your portfolio is the most important thing for getting a job. Show your creative process, not just final designs.",
			Duration: "5 years", Links: []roadmapLink{
				{Title: "TU Institute of Engineering - B.Arch program", URL: "https://www.ioe.edu.np"},
				{Title: "Kathmandu University - Architecture", URL: "https://www.ku.edu.np"},
				{Title: "Architecture design studio tips (YouTube)", URL: "https://www.youtube.com/results?search_query=architecture+design+studio+tips"},
			}},
			{StepNumber: 3, Title: "Complete internship and get licensed", Description: "After graduation, complete a mandatory internship (typically 1-2 years) under a registered architect. Work on real projects — residential houses, commercial buildings, schools. Learn how designs become reality. Understand construction documentation, cost estimation, and client management. Then register with the Nepal Engineering Council as a licensed architect. The licensing process includes submitting your portfolio, internship records, and passing an exam. A licensed architect can independently design buildings, submit plans for permits, and supervise construction.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "NEC - Architect licensing process", URL: "https://www.nec.gov.np"},
				{Title: "Architecture internship tips (YouTube)", URL: "https://www.youtube.com/results?search_query=architecture+internship+nepal"},
				{Title: "SONA - Young architects forum", URL: "https://www.sonanepal.com"},
			}},
			{StepNumber: 4, Title: "Work in an architecture firm to build experience", Description: "Join an architecture firm in Kathmandu or Pokhara. Start as a junior architect working on design development, construction drawings, and 3D visualization. Learn from senior architects. Work on different project types — residential, commercial, institutional, heritage conservation. Build your portfolio with completed projects. Develop skills in project management, client communication, and contractor coordination. Architecture is a profession where experience matters enormously. Each project teaches you something new about design, materials, and people.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Find architecture jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Architecture firms in Kathmandu", URL: "https://www.sonanepal.com"},
				{Title: "Building your architecture portfolio (YouTube)", URL: "https://www.youtube.com/results?search_query=architecture+portfolio+tips"},
			}},
			{StepNumber: 5, Title: "Specialize in earthquake-resistant or sustainable design", Description: "Nepal is earthquake-prone. Specializing in earthquake-resistant design is valuable. Learn about base isolation, shear walls, and ductile detailing. Sustainable design (green buildings, passive solar, local materials) is growing in importance. Get certified in LEED or similar green building standards. Specializing in heritage conservation (preserving traditional architecture) is unique to Nepal and has growing demand. Specialization makes you stand out and allows you to charge higher fees.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Earthquake-resistant architecture training (YouTube)", URL: "https://www.youtube.com/results?search_query=earthquake+resistant+architecture+nepal"},
				{Title: "Sustainable design and green building Nepal", URL: "https://www.youtube.com/results?search_query=sustainable+architecture+nepal"},
				{Title: "Heritage conservation architecture Nepal", URL: "https://www.sonanepal.com"},
			}},
			{StepNumber: 6, Title: "Start your own practice or become a design leader", Description: "Experienced architects can start their own firm. Building a practice takes time — start with small residential projects, build a reputation, take on larger projects. Some architects become design directors at large firms. Others teach at architecture colleges. Some specialize in high-end resorts and tourism architecture — a growing field in Nepal. Architecture is a profession where you can see your work everywhere — in the buildings that shape Nepal's cities and towns. Your designs will outlive you. That is a legacy few professions can match.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting an architecture firm in Nepal", URL: "https://www.sonanepal.com"},
				{Title: "Resort and tourism architecture Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=resort+architecture+nepal"},
				{Title: "Teaching architecture in Nepal", URL: "https://www.ioe.edu.np"},
			}},
		},
	}
}

func quantitySurveyor() careerSeed {
	return careerSeed{
		CategoryName: "Engineering & Construction", CategorySlug: "engineering-construction", CategoryIcon: "📐",
		Title: "Quantity Surveyor", Slug: "quantity-surveyor",
		Summary: "Quantity surveyors manage costs on construction projects. They prepare budgets, measure quantities, value completed work, and ensure projects are financially controlled.",
		Description: "A quantity surveyor (QS) manages all costs related to construction projects. They prepare cost estimates, measure quantities of materials, value completed work for payments, and help control project budgets. In Nepal, quantity surveyors work for construction companies, government infrastructure projects, hydropower developers, and consulting firms. The role is critical because construction projects often run over budget. QS professionals help keep costs under control. The Nepal Institution of Quantity Surveyors (NIQS) is the professional body. Quantity surveyors need knowledge of construction methods, measurement standards (CESMM, POMI), contract law, and cost management. It is a well-paid profession with growing demand as Nepal invests in infrastructure.",
		DailyTasks: []string{"Prepare cost estimates and budgets for construction projects", "Measure and calculate quantities of materials from drawings", "Prepare bills of quantities (BOQ) for tenders", "Evaluate contractor payment applications and variations", "Monitor project costs and report on budget status", "Assist in contract administration and procurement", "Value completed work for progress payments"},
		Skills: []string{"Cost estimation and budgeting", "Measurement and quantification from drawings", "Knowledge of standard methods of measurement", "Contract administration (FIDIC, Nepal PWD conditions)", "Understanding of construction methods and materials", "Analytical and numerical skills", "Proficiency in MS Excel and QS software"},
		SalaryMin: 350000, SalaryMax: 1800000, Difficulty: 4, FutureProof: 75,
		EducationReq: "Bachelor's in Quantity Surveying or Civil Engineering with QS specialization. Diploma in QS from CTEVT. Membership in Nepal Institution of Quantity Surveyors (NIQS) valued. Professional certifications like MRICS for international work.",
		Outlook: "Infrastructure development in Nepal is increasing demand for QS professionals. Hydropower, roads, airports, and building construction all need cost management. The profession is growing and well-recognized. Experienced QS professionals can work internationally in the Middle East, Australia, and the UK where the profession is well-established.",
		Tags: []string{"engineering", "construction", "cost-management", "finance", "contracts"},
		Resources: []resourceSeed{
			{Title: "Nepal Institution of Quantity Surveyors", URL: "https://www.niqs.org.np", Description: "Professional body for quantity surveyors in Nepal"},
			{Title: "Department of Urban Development Nepal", URL: "https://www.dudbc.gov.np", Description: "Government construction standards and cost data"},
			{Title: "Merojob Construction Jobs", URL: "https://www.merojob.com", Description: "Find quantity surveyor jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Build math and technical drawing skills", Description: "Quantity surveying requires strong math skills (measurement, calculation, percentages) and ability to read technical drawings. Focus on math, physics, and drawing in SEE and +2. Read about construction and how buildings are built. If you enjoy working with numbers and details and want to work in construction without being on-site all the time, QS could be the right path for you. Accuracy is everything in QS — a small error can cost millions.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "NIQS - Quantity surveying career info", URL: "https://www.niqs.org.np"},
				{Title: "What is quantity surveying? (YouTube)", URL: "https://www.youtube.com/results?search_query=quantity+surveying+for+beginners"},
				{Title: "Construction technology basics", URL: "https://www.youtube.com/results?search_query=construction+technology+nepal"},
			}},
			{StepNumber: 2, Title: "Complete QS education and training", Description: "Enroll in a Quantity Surveying degree or diploma. Study measurement, cost estimation, contract law, construction technology, and procurement. Learn measurement standards and software. Do internships with construction companies or consulting firms. Practical experience in measuring quantities from real drawings and preparing cost estimates is essential. Your education gives you the foundation but real learning happens on the job. Pay attention to details — every beam, every cubic meter of concrete, every kilogram of reinforcement must be counted.",
			Duration: "3-4 years", Links: []roadmapLink{
				{Title: "QS programs in Nepal (NIQS)", URL: "https://www.niqs.org.np"},
				{Title: "CESMM measurement explained (YouTube)", URL: "https://www.youtube.com/results?search_query=cesmm+measurement+guide"},
				{Title: "Construction cost estimation software", URL: "https://www.youtube.com/results?search_query=construction+cost+estimation+software"},
			}},
			{StepNumber: 3, Title: "Join a construction project and learn on the job", Description: "Work as a junior QS on a construction site or in a consultancy office. Learn to prepare bills of quantities, evaluate contractor payments, and track project costs. Understand how construction works in practice — what materials are used, how labor is organized, how progress is measured. Build relationships with contractors, engineers, and project managers. Communication is important in QS — you need to discuss costs clearly with people who may not have financial backgrounds. Your accurate work helps projects stay on budget and avoids disputes.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "Find QS jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Junior quantity surveyor tips (YouTube)", URL: "https://www.youtube.com/results?search_query=junior+quantity+surveyor+tips"},
				{Title: "FIDIC contracts for Nepali construction", URL: "https://www.youtube.com/results?search_query=fidic+contract+nepal"},
			}},
			{StepNumber: 4, Title: "Get professional membership with NIQS", Description: "Join the Nepal Institution of Quantity Surveyors as a member. NIQS provides professional development, networking, and recognition. Attend NIQS events and workshops. Connect with other QS professionals. Professional membership adds credibility and helps in career advancement. Work toward chartered status if you plan to work internationally. MRICS (Royal Institution of Chartered Surveyors) is recognized worldwide and allows you to work in many countries.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "NIQS - Membership information", URL: "https://www.niqs.org.np"},
				{Title: "MRICS certification pathway", URL: "https://www.rics.org"},
				{Title: "NIQS training and events", URL: "https://www.niqs.org.np"},
			}},
			{StepNumber: 5, Title: "Specialize in a sector or take on larger projects", Description: "Specialize in a construction sector: hydropower (large budgets, complex contracts), building construction (residential, commercial), roads and bridges, or industrial projects. Each sector has different measurement standards and contract conditions. Specialization makes you more valuable. Take on larger projects with bigger budgets and more complexity. Senior QS professionals manage cost planning for billion-rupee projects. They lead teams of junior QS staff. Develop your leadership and communication skills alongside technical expertise.",
			Duration: "3-5 years", Links: []roadmapLink{
				{Title: "Hydropower quantity surveying (YouTube)", URL: "https://www.youtube.com/results?search_query=hydropower+quantity+surveying"},
				{Title: "Building construction cost planning", URL: "https://www.niqs.org.np"},
				{Title: "Senior QS role and responsibilities", URL: "https://www.youtube.com/results?search_query=senior+quantity+surveyor+role"},
			}},
			{StepNumber: 6, Title: "Become a cost consultant or start your own practice", Description: "Experienced QS professionals can become independent cost consultants, advising clients on project feasibility, cost planning, and contract administration. Some start their own QS consultancy firms. Others work for international development banks as cost experts on infrastructure projects. Senior QS roles offer excellent salaries — some of the highest in the construction industry. Quantity surveyors are the financial guardians of construction projects. Your work ensures that Nepal's infrastructure investments deliver value for money. Every bridge built, every road paved, every hydropower plant constructed — your cost management helped make it happen.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a QS consultancy in Nepal", URL: "https://www.niqs.org.np"},
				{Title: "Cost consultant career path (YouTube)", URL: "https://www.youtube.com/results?search_query=cost+consultant+career+nepal"},
				{Title: "International QS opportunities", URL: "https://www.rics.org"},
			}},
		},
	}
}

func constructionManager() careerSeed {
	return careerSeed{
		CategoryName: "Engineering & Construction", CategorySlug: "engineering-construction", CategoryIcon: "🏗️",
		Title: "Construction Manager", Slug: "construction-manager",
		Summary: "Construction managers oversee construction projects from start to finish. They manage schedules, budgets, workers, materials, and quality on building and infrastructure sites.",
		Description: "A construction manager plans, directs, and coordinates construction projects. They are responsible for completing projects on time, within budget, and according to specifications. Construction managers work on building projects (houses, apartments, commercial buildings), infrastructure projects (roads, bridges, water systems), and industrial projects (hydropower plants, factories). They manage teams of workers, subcontractors, suppliers, and equipment. The job requires knowledge of construction methods, project management, safety regulations, and contract administration. In Nepal, construction managers are needed for the country's growing infrastructure and building sectors. The job involves working on-site, often in challenging conditions. It is a high-responsibility role with correspondingly good pay.",
		DailyTasks: []string{"Plan and schedule construction activities and timelines", "Manage site workers, subcontractors, and suppliers", "Monitor construction progress and quality", "Control project costs and manage the budget", "Ensure health and safety standards are followed", "Solve problems and make decisions to keep work moving", "Report progress to clients and senior management"},
		Skills: []string{"Construction project planning and scheduling", "Team leadership and people management", "Budget and cost control", "Knowledge of construction methods and materials", "Safety management and risk assessment", "Contract administration and negotiation", "Problem-solving and decision-making under pressure"},
		SalaryMin: 500000, SalaryMax: 3000000, Difficulty: 4, FutureProof: 78,
		EducationReq: "Bachelor's in Civil Engineering or Construction Management. Master's in Construction Management preferred for senior roles. PMP or equivalent project management certification valuable. Extensive site experience essential.",
		Outlook: "Nepal's construction sector is growing with infrastructure projects, hydropower development, and urban construction. Experienced construction managers are in high demand and short supply. The role offers excellent salaries and career progression. Government infrastructure projects and hydropower development offer the largest opportunities.",
		Tags: []string{"engineering", "construction", "management", "high-salary", "leadership"},
		Resources: []resourceSeed{
			{Title: "Construction Association of Nepal", URL: "https://www.canepal.org", Description: "Professional body for construction companies and managers"},
			{Title: "Department of Roads Nepal", URL: "https://www.dor.gov.np", Description: "Major infrastructure projects and construction standards"},
			{Title: "Merojob Construction Jobs", URL: "https://www.merojob.com", Description: "Find construction manager jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Get construction experience from the ground up", Description: "Work on a construction site in any role — laborer, mason helper, site supervisor. Understand how construction actually happens. Learn the practical aspects: how concrete is poured, how rebar is tied, how walls are built. Site experience is invaluable for a construction manager. You cannot manage what you do not understand. Many successful construction managers started as laborers and worked their way up. Respect the workers who build structures with their hands — they have knowledge you will never get from a book.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Construction Association of Nepal - Training", URL: "https://www.canepal.org"},
				{Title: "Construction site basics (YouTube)", URL: "https://www.youtube.com/results?search_query=construction+site+work+nepal"},
				{Title: "Building construction methods explained", URL: "https://www.youtube.com/results?search_query=building+construction+methods+nepal"},
			}},
			{StepNumber: 2, Title: "Get education in civil engineering or construction management", Description: "Earn a degree in Civil Engineering or Construction Management. Study project management, construction methods, contract administration, safety management, and cost estimation. Learn scheduling software (MS Project, Primavera). Understand quality control and testing procedures. Your education provides the theoretical foundation for management decisions. Combine classroom learning with continued site work during holidays. The best construction managers have both theoretical knowledge and practical experience.",
			Duration: "3-4 years", Links: []roadmapLink{
				{Title: "Construction Management programs Nepal", URL: "https://www.ioe.edu.np"},
				{Title: "MS Project for construction scheduling (YouTube)", URL: "https://www.youtube.com/results?search_query=ms+project+construction+scheduling"},
				{Title: "Construction safety management Nepal", URL: "https://www.youtube.com/results?search_query=construction+safety+nepal"},
			}},
			{StepNumber: 3, Title: "Work as a site supervisor or assistant project manager", Description: "Join a construction company as a site supervisor, project engineer, or assistant project manager. Manage day-to-day site operations. Learn to read drawings, manage workers, order materials, and track progress. Understand how projects are planned and how plans change when reality intervenes. Build relationships with subcontractors and suppliers. Learn to communicate effectively with clients, architects, and workers. Take responsibility for small sections of the project first. Prove you can deliver before taking on larger responsibility.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Find construction management jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Site supervisor responsibilities (YouTube)", URL: "https://www.youtube.com/results?search_query=site+supervisor+role+nepal"},
				{Title: "Construction project coordination tips", URL: "https://www.youtube.com/results?search_query=construction+coordination+skills"},
			}},
			{StepNumber: 4, Title: "Get PMP certification and advance your knowledge", Description: "Project Management Professional (PMP) certification is highly valued for construction managers. Study for the PMP exam through training institutes in Nepal. Learn earned value management, risk management, quality management, and leadership. Attend construction management workshops and seminars. PMP certification demonstrates your commitment to professional standards and your ability to manage complex projects. Many large projects require PMP-certified managers. The certification opens doors to better positions and higher pay.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "PMP certification in Nepal", URL: "https://www.pmi.org"},
				{Title: "PMP exam preparation tips (YouTube)", URL: "https://www.youtube.com/results?search_query=pmp+exam+preparation+nepal"},
				{Title: "Construction project management best practices", URL: "https://www.youtube.com/results?search_query=construction+project+management+best+practices"},
			}},
			{StepNumber: 5, Title: "Take on larger and more complex projects", Description: "Move from managing small buildings to large infrastructure projects. Manage a team of engineers, supervisors, and administrative staff. Handle project budgets of 10 crore rupees and above. Learn to manage multiple projects simultaneously. Develop expertise in a sector: hydropower construction (most complex), road and bridge construction, high-rise buildings, or industrial projects. Each sector has unique challenges. Large project experience is what separates good construction managers from great ones. The challenges are bigger but the satisfaction of completing a major project is immense.",
			Duration: "3-6 years", Links: []roadmapLink{
				{Title: "Hydropower construction management (YouTube)", URL: "https://www.youtube.com/results?search_query=hydropower+construction+management+nepal"},
				{Title: "Managing large construction projects", URL: "https://www.youtube.com/results?search_query=large+construction+project+management"},
				{Title: "Construction claims and dispute resolution", URL: "https://www.canepal.org"},
			}},
			{StepNumber: 6, Title: "Become a director or start your own construction company", Description: "Top construction managers become project directors, operations directors, or managing directors of construction companies. Some start their own construction firm. Starting a construction company requires capital, licenses (from the Department of Urban Development), a track record, and a network. The potential rewards are significant. Nepal needs experienced construction managers to deliver quality infrastructure. Every building, road, and dam you manage is a contribution to Nepal's development. Your leadership turns designs into reality and creates assets that serve people for generations.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a construction company in Nepal", URL: "https://www.canepal.org"},
				{Title: "Construction company licensing Nepal", URL: "https://www.dudbc.gov.np"},
				{Title: "Construction industry leadership (YouTube)", URL: "https://www.youtube.com/results?search_query=construction+leadership+nepal"},
			}},
		},
	}
}

func surveyor() careerSeed {
	return careerSeed{
		CategoryName: "Engineering & Construction", CategorySlug: "engineering-construction", CategoryIcon: "📏",
		Title: "Surveyor", Slug: "surveyor",
		Summary: "Surveyors measure land and create maps for construction, property boundaries, and infrastructure projects. Their measurements are the starting point for all construction.",
		Description: "A surveyor measures and maps land for construction projects, property boundaries, and infrastructure development. They use specialized equipment like total stations, GPS, and drones to take precise measurements. Surveyors in Nepal are essential for every construction project — from building a house to building a highway or hydropower plant. They work for the Survey Department of Nepal, construction companies, land administration offices, and private surveying firms. Surveyors also handle land registration, boundary disputes, and topographic mapping. The Survey Department of Nepal manages the national survey system. Surveying is a well-respected profession with good demand as Nepal develops infrastructure and manages land records.",
		DailyTasks: []string{"Visit sites to measure land using surveying equipment", "Use GPS, total station, and level to take measurements", "Create maps and drawings from survey data", "Determine property boundaries for registration or disputes", "Provide survey data for construction projects", "Process survey data using CAD and GIS software", "Prepare survey reports and documentation"},
		Skills: []string{"Operation of surveying equipment (total station, GPS, level)", "Knowledge of surveying methods and calculations", "Map reading and creation using GIS software", "Understanding of land registration and cadastral systems", "Computer skills (AutoCAD, ArcGIS, QGIS)", "Physical fitness for fieldwork in challenging terrain", "Accuracy and attention to detail"},
		SalaryMin: 300000, SalaryMax: 1200000, Difficulty: 3, FutureProof: 70,
		EducationReq: "Bachelor's in Geomatics Engineering or Surveying. Diploma in Surveying from CTEVT. Registration with Nepal Engineering Council (for engineers) or Survey Department. Government surveyor positions through Loksewa exam.",
		Outlook: "Surveyors are needed for every construction project and land transaction. Nepal's land administration modernization, infrastructure development, and hydropower projects all require surveyors. The Survey Department regularly recruits surveyors. Private sector demand is growing. Drone surveying is creating new opportunities and methods.",
		Tags: []string{"engineering", "surveying", "outdoor", "measurement", "construction"},
		Resources: []resourceSeed{
			{Title: "Survey Department Nepal", URL: "https://www.dos.gov.np", Description: "National survey authority and career information"},
			{Title: "Nepal Engineering Council - Geomatics", URL: "https://www.nec.gov.np", Description: "Licensing for geomatics/surveying engineers"},
			{Title: "Merojob Engineering Jobs", URL: "https://www.merojob.com", Description: "Find surveyor jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn math, geography, and map reading", Description: "Surveying requires good math skills (trigonometry, geometry, algebra) and understanding of maps and coordinates. Study these subjects in SEE and +2. Learn to read topographic maps and understand scale, contours, and coordinates. Practice using a compass and measuring distances. If you enjoy working outdoors, being precise, and seeing your work used in construction projects, surveying could be the right career. Spend time observing surveyors at work if you can. Their measurements are the foundation of every building and road.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Survey Department Nepal - Careers", URL: "https://www.dos.gov.np"},
				{Title: "Surveying career overview (YouTube)", URL: "https://www.youtube.com/results?search_query=surveying+career+nepal"},
				{Title: "Map reading basics for beginners", URL: "https://www.youtube.com/results?search_query=map+reading+skills+beginners"},
			}},
			{StepNumber: 2, Title: "Complete surveying education and training", Description: "Enroll in a Diploma in Surveying (CTEVT) or Bachelor's in Geomatics Engineering (TU, KU). Study surveying methods, geodesy, cartography, photogrammetry, GIS, and land registration. Gain practical experience with surveying equipment during fieldwork. Learn to use total station, GPS, and digital levels. Learn CAD software for drafting survey maps. Surveying education combines classroom theory with extensive fieldwork. Being comfortable outdoors in all weather is essential. Accuracy is critical — in surveying, small errors can cause big problems.",
			Duration: "3-4 years", Links: []roadmapLink{
				{Title: "CTEVT - Diploma in Surveying", URL: "https://www.ctevt.org.np"},
				{Title: "TU Geomatics Engineering program", URL: "https://www.ioe.edu.np"},
				{Title: "Total station training for surveyors (YouTube)", URL: "https://www.youtube.com/results?search_query=total+station+training+for+beginners"},
			}},
			{StepNumber: 3, Title: "Start working and gaining field experience", Description: "Join a surveying firm, construction company, or the Survey Department. Work as a junior surveyor. Assist senior surveyors in field measurements and data processing. Learn to work in different terrains — flat land, hills, mountains, forests. Experience in Nepal's diverse terrain makes you a versatile surveyor. Learn about the land registration process and cadastral surveying. Build speed and accuracy in your work. Every survey you complete adds to your experience and confidence. Learn to deal with landowners and the public — surveying often involves explaining boundaries to people.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Find surveyor jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Survey Department - Recruitment", URL: "https://www.dos.gov.np"},
				{Title: "Cadastral surveying in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=cadastral+surveying+nepal"},
			}},
			{StepNumber: 4, Title: "Learn GIS and modern surveying technologies", Description: "GIS (Geographic Information Systems) is increasingly important in surveying. Learn QGIS (free) or ArcGIS to create digital maps and analyze spatial data. Learn drone surveying — drones can survey large areas quickly. Drone surveyors are in high demand for mining, construction, and agriculture. Get certified in drone piloting (CAAN license). Modern surveyors who combine traditional skills with GIS and drone technology are the most valuable. The field is evolving rapidly and lifelong learning is essential.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "QGIS training for surveyors (YouTube)", URL: "https://www.youtube.com/results?search_query=qgis+for+beginners+surveyors"},
				{Title: "Drone surveying and mapping training", URL: "https://www.youtube.com/results?search_query=drone+surveying+nepal"},
				{Title: "ArcGIS for surveyors", URL: "https://www.youtube.com/results?search_query=arcgis+surveying+tutorial"},
			}},
			{StepNumber: 5, Title: "Specialize in a survey type or get licensed", Description: "Specialize in: cadastral surveying (land boundaries), engineering surveying (construction), topographic surveying (mapping), hydrographic surveying (water bodies), or geodetic surveying (large-scale accurate measurements). Each specialty requires specific knowledge and skills. Licensed surveyors (registered with Survey Department) can independently certify surveys for land registration. Licensure requires passing exams and demonstrating experience. Licensed surveyors earn more and have more professional autonomy. They can take legal responsibility for boundary determinations.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Survey Department - Licensed surveyor path", URL: "https://www.dos.gov.np"},
				{Title: "Engineering surveying specialization (YouTube)", URL: "https://www.youtube.com/results?search_query=engineering+surveying+nepal"},
				{Title: "Land registration and cadastral survey guide", URL: "https://www.dos.gov.np"},
			}},
			{StepNumber: 6, Title: "Lead survey teams or start your own practice", Description: "Experienced surveyors lead teams on major projects — supervising multiple survey crews, managing data processing, and delivering final surveys. Some open their own survey practice or firm. Starting a survey firm requires equipment investment (total station, GPS, computer systems) and licensing. Government survey projects are often contracted to private firms. Surveyors also work as expert witnesses in land disputes. Every construction project starts with a surveyor's measurements. Your precise work is the foundation that all construction builds upon. Without surveyors, nothing gets built in the right place.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a survey firm in Nepal", URL: "https://www.dos.gov.np"},
				{Title: "Survey project management (YouTube)", URL: "https://www.youtube.com/results?search_query=survey+project+management+nepal"},
				{Title: "Expert witness surveying career", URL: "https://www.youtube.com/results?search_query=surveyor+expert+witness"},
			}},
		},
	}
}

func journalist() careerSeed {
	return careerSeed{
		CategoryName: "Media & Arts", CategorySlug: "media-arts", CategoryIcon: "🎤",
		Title: "Journalist", Slug: "journalist",
		Summary: "Journalists investigate stories, interview people, and write news articles that inform the public about current events.",
		Description: "A journalist researches and reports news for newspapers, TV, radio, or online media. They cover politics, social issues, sports, entertainment, business, and more. Journalists in Nepal work for publications like The Kathmandu Post, Himalayan Times, Nepali Times, Kantipur, as well as TV channels such as Kantipur TV, Nepal 1, AP1, and online news portals. Journalists must be curious, persistent, and committed to truth. They interview people, verify facts, and write stories that help the public understand what is happening. Nepal's media landscape has grown rapidly, with many opportunities for journalists who uphold ethical standards and produce quality reporting. The profession requires courage, especially when covering sensitive political or social issues.",
		DailyTasks: []string{"Research news stories by reading, talking to sources, and monitoring events", "Interview people — politicians, activists, experts, ordinary citizens", "Write news articles, features, or investigative reports", "Edit and fact-check stories before publication", "Attend press conferences, seminars, and events", "Follow up on stories to track developments", "Produce content for social media to reach wider audiences"},
		Skills: []string{"Strong writing and grammar in English and/or Nepali", "Interviewing skills and ability to talk to anyone", "Research and fact-checking abilities", "Understanding of media ethics and libel laws", "Photography or video skills (basic)", "Social media management", "Time management and deadline orientation", "Courage and persistence for investigative journalism"},
		SalaryMin: 240000, SalaryMax: 1200000, Difficulty: 3, FutureProof: 55,
		EducationReq: "Bachelor's in Journalism, Mass Communication, or related field. Master's in Journalism (KU, TU) for senior positions. Courses from Nepal Journalism Institute or Nepal College of Journalism.",
		Outlook: "Nepal's media is growing with new online news portals and TV channels. Quality journalism is in demand. Digital journalism and multimedia skills are increasingly important. Freelance journalism offers flexibility. Investigative journalism has high impact but low pay. Many journalists leave the profession due to low pay and pressure.",
		Tags: []string{"media", "writing", "communication", "news", "investigation"},
		Resources: []resourceSeed{
			{Title: "The Kathmandu Post", URL: "https://www.kathmandupost.com", Description: "Leading English-language daily in Nepal"},
			{Title: "Merojob Media Jobs", URL: "https://www.merojob.com", Description: "Find journalism jobs in Nepal"},
			{Title: "Nepal Press Institute", URL: "https://www.npi.com.np", Description: "Journalism training and resources in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Read and write constantly to develop your skills", Description: "Journalism starts with being a good writer. Read newspapers, magazines, and online articles every day — in both Nepali and English. Notice how different writers structure stories, use quotes, and explain complex topics. Start a blog or write for a school/college newspaper. Practice writing news articles, features, interviews, and opinion pieces. The more you write, the better you become. Read The Kathmandu Post and Nepali Times daily. Listen to news on BBC Nepali and Radio Nepal. Good journalists are also good readers and thinkers.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "The Kathmandu Post Online", URL: "https://www.kathmandupost.com"},
				{Title: "BBC Nepali News", URL: "https://www.bbc.com/nepali"},
				{Title: "Journalism writing tips for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=journalism+writing+tips+for+beginners"},
			}},
			{StepNumber: 2, Title: "Get formal journalism education", Description: "Enroll in a Bachelor's in Journalism or Mass Communication. Study news writing, media law, ethics, feature writing, and broadcast journalism. Learn about the history of journalism in Nepal and the role of press freedom. Participate in college media — radio, newspaper, or online platforms. A good journalism program will give you the theoretical foundation and ethics training essential for the profession. Internships are often part of the program. Build relationships with teachers who are experienced journalists. They can guide you and connect you with media houses.",
			Duration: "3-4 years", Links: []roadmapLink{
				{Title: "KU Journalism Program", URL: "https://www.ku.edu.np"},
				{Title: "Journalism ethics and responsibility guide", URL: "https://www.youtube.com/results?search_query=journalism+ethics+nepal"},
				{Title: "Mass Communication programs in Nepal", URL: "https://www.edusanjal.com"},
			}},
			{StepNumber: 3, Title: "Start as a trainee or junior reporter", Description: "Join a newspaper, TV station, or online news portal as a trainee or junior reporter. Accept that you will start with small stories and beat reporting (crime, local events, civic issues). Work hard, meet deadlines, and build your portfolio. Every story you publish adds to your experience and reputation. Learn from senior editors and reporters. They will teach you the practical aspects of journalism that college cannot. Be willing to work long hours and cover stories on weekends. Persistence is the most important quality for a journalist.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Find journalism jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Tips for junior reporters (YouTube)", URL: "https://www.youtube.com/results?search_query=junior+reporter+tips+nepal"},
				{Title: "How newsrooms work in Nepal", URL: "https://www.youtube.com/results?search_query=nepal+newsroom+work"},
			}},
			{StepNumber: 4, Title: "Develop expertise in a beat or specialty", Description: "Specialize in a reporting beat — politics, education, health, environment, business, sports, or investigative journalism. Subject matter expertise makes you more valuable. Learn to develop sources — people who trust you and share information. Sources are a journalist's most important asset. If you are interested in investigative journalism, learn data journalism, document analysis, and how to handle sensitive sources. A specialized journalist earns more and is harder to replace.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "Investigative journalism resource center (Nepal)", URL: "https://www.youtube.com/results?search_query=investigative+journalism+nepal"},
				{Title: "Data journalism for beginners", URL: "https://www.youtube.com/results?search_query=data+journalism+for+beginners"},
				{Title: "Political reporting in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=political+reporting+nepal"},
			}},
			{StepNumber: 5, Title: "Become a senior reporter, editor, or media manager", Description: "Experienced journalists become senior reporters, bureau chiefs, editors, or news managers. Editors guide the newsroom, assign stories, and ensure quality. Some journalists move into media management or public relations. Others teach journalism at universities. Many successful journalists in Nepal have moved from reporting to editing, media entrepreneurship, or communication roles in NGOs and international organizations. Your journalism skills are valuable beyond the newsroom.",
			Duration: "3-5 years", Links: []roadmapLink{
				{Title: "Nepal journalists association - career development", URL: "https://www.fnjnepal.org"},
				{Title: "Transitioning from reporter to editor (YouTube)", URL: "https://www.youtube.com/results?search_query=reporter+to+editor+career+transition"},
				{Title: "Media management in Nepal", URL: "https://www.youtube.com/results?search_query=media+management+nepal"},
			}},
			{StepNumber: 6, Title: "Become a media entrepreneur or pursue advanced journalism", Description: "Some top journalists start their own media ventures — online news portals, news agencies, or production houses. Others pursue international journalism with UN agencies, BBC, or international media. Some study abroad for master's degrees in journalism. Media entrepreneurship offers freedom but comes with risk. International journalism offers higher pay and exposure. Journalists who have built a strong reputation can move into diplomacy, politics, or international development. Your voice as a journalist can shape public opinion and bring positive change to society.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a media startup in Nepal", URL: "https://www.merojob.com"},
				{Title: "International journalism career paths", URL: "https://www.youtube.com/results?search_query=international+journalism+career"},
				{Title: "Press freedom and journalism advocacy", URL: "https://www.youtube.com/results?search_query=press+freedom+nepal"},
			}},
		},
	}
}

func radioJockey() careerSeed {
	return careerSeed{
		CategoryName: "Media & Arts", CategorySlug: "media-arts", CategoryIcon: "📻",
		Title: "Radio Jockey (RJ)", Slug: "radio-jockey",
		Summary: "RJ hosts radio shows, plays music, talks to listeners, and creates entertaining or informative radio content.",
		Description: "A Radio Jockey (RJ) hosts radio programs on FM stations. They play music, chat with listeners, share news, conduct contests, interview guests, and keep the audience engaged. Nepal has over 500 FM radio stations, making it one of the most vibrant radio landscapes in the world. RJs are the voice of these stations. Good RJs have a pleasant voice, quick wit, good knowledge of music, and the ability to engage listeners. Radio is still the most accessible mass medium in Nepal, especially in rural areas. RJs can become celebrities with large fan followings. Many successful RJs have also moved into TV hosting, events, and voice-over work.",
		DailyTasks: []string{"Prepare show content — music selection, topics, scripts", "Host live radio shows — speak, play music, talk to listeners", "Take listener calls and interact with the audience", "Interview guests — musicians, celebrities, experts", "Promote the station on social media", "Attend station meetings and promotional events", "Record voice-overs, promos, and advertisements"},
		Skills: []string{"Excellent verbal communication and voice modulation", "Good music knowledge across genres", "Quick thinking and ability to handle live situations", "Interviewing and conversation skills", "Social media engagement skills", "Basic audio editing (Adobe Audition, Audacity)", "Creativity and show planning", "Energy and positive attitude on air"},
		SalaryMin: 240000, SalaryMax: 900000, Difficulty: 2, FutureProof: 45,
		EducationReq: "Bachelor's in Mass Communication or Journalism is preferred. Voice training and RJ courses from institutes like Nepal College of Journalism or Sound of Nepal. Good English and Nepali language skills essential.",
		Outlook: "Nepal has hundreds of FM stations but many operate with small teams. Competition for top RJ positions in Kathmandu is high. Podcasting is creating new opportunities for audio content creators. Many RJs also work as event hosts, voice-over artists, and content creators. Radio faces competition from digital media but remains relevant in rural areas.",
		Tags: []string{"media", "radio", "entertainment", "communication", "music"},
		Resources: []resourceSeed{
			{Title: "Association of Community Radio Broadcasters Nepal", URL: "https://www.acorb.org.np", Description: "Network of community radio stations in Nepal"},
			{Title: "Nepal FM radio directory", URL: "https://www.merojob.com", Description: "List of FM stations and radio jobs in Nepal"},
			{Title: "Voice training and RJ courses Nepal", URL: "https://www.youtube.com/results?search_query=radio+jockey+training+nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Develop your voice and communication skills", Description: "Your voice is your instrument. Practice speaking clearly, with energy and emotion. Record yourself speaking and listen carefully — improve your pronunciation, pace, and modulation. Read aloud from newspapers and books. Practice in both Nepali and English. Learn to speak naturally while also being entertaining. Listen to popular RJs on Nepali FM stations. Notice how they start shows, handle calls, and transition between segments. Good RJs sound like they are talking to one person, not a crowd. That intimate connection is the secret to radio.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Voice modulation exercises (YouTube)", URL: "https://www.youtube.com/results?search_query=voice+modulation+exercises+for+RJ"},
				{Title: "Listen to Nepali FM radio online", URL: "https://www.radionepal.gov.np"},
				{Title: "Public speaking tips for beginners", URL: "https://www.youtube.com/results?search_query=public+speaking+tips+for+beginners+nepal"},
			}},
			{StepNumber: 2, Title: "Get training and education in radio", Description: "Enroll in an RJ training program or Mass Communication degree. Learn about radio production, audio editing, show planning, and broadcast ethics. Study the technical aspects — mixing boards, microphones, recording software. Practice creating demo recordings (air checks) that showcase your voice and style. Your demo is your resume in the radio industry. Make multiple demos showing different styles — music show, talk show, news reading. The quality of your demo determines whether a station will give you a chance. Invest in good equipment for recording at home.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Mass Communication programs at KU", URL: "https://www.ku.edu.np"},
				{Title: "How to create a radio demo (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+make+a+radio+demo"},
				{Title: "Audio editing with Audacity tutorial", URL: "https://www.youtube.com/results?search_query=audacity+audio+editing+tutorial"},
			}},
			{StepNumber: 3, Title: "Get your first break at a radio station", Description: "Apply for internships or trainee positions at FM stations. Be willing to start with non-air shifts — assisting producers, screening calls, running the board. Many RJs started as interns and worked their way to on-air positions. Once you get an on-air shift, record every show and listen back to improve. Ask senior RJs for feedback. Build relationships with station staff. Radio is a team effort. The more reliable and talented you are, the better shifts you will get. Be prepared for early morning or late night shifts — that is how most RJs start.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Find radio jobs in Nepal (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Tips for new radio jockeys (YouTube)", URL: "https://www.youtube.com/results?search_query=tips+for+new+radio+jockeys"},
				{Title: "ACORB radio training resources", URL: "https://www.acorb.org.np"},
			}},
			{StepNumber: 4, Title: "Develop your niche and grow your audience", Description: "Great RJs have a unique style that listeners love. Develop your niche — music expert, comedy, talk show host, or social issues. Build a loyal audience by being consistent and authentic. Use social media to promote your show and interact with listeners. Create content beyond your radio show — podcasts, YouTube videos, event hosting. The most successful RJs in Nepal are known beyond their radio shows — they are public personalities. Learn about digital content creation to stay relevant. Radio alone is not enough for career growth anymore.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Podcasting for RJs (YouTube)", URL: "https://www.youtube.com/results?search_query=podcasting+for+radio+jockeys"},
				{Title: "Social media marketing for radio personalities", URL: "https://www.youtube.com/results?search_query=social+media+for+radio+shows"},
				{Title: "Event hosting tips for RJs", URL: "https://www.youtube.com/results?search_query=event+hosting+skills+for+RJ"},
			}},
			{StepNumber: 5, Title: "Expand into TV, events, or voice-over work", Description: "Experienced RJs often expand into TV hosting, event emceeing, voice-over for advertisements, and corporate video narration. These opportunities pay significantly better than radio alone. Build a portfolio of voice samples for different purposes — commercial voice-over, documentary narration, event hosting. Network with advertising agencies and production houses. Your voice is a valuable asset — use it across multiple platforms. Voice-over work for ads pays well per project. Many top Nepali voice artists started as RJs.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Voice-over career guide (YouTube)", URL: "https://www.youtube.com/results?search_query=voice+over+career+guide+nepal"},
				{Title: "TV hosting tips from Nepali TV hosts", URL: "https://www.youtube.com/results?search_query=tv+hosting+nepal"},
				{Title: "Event emcee training (YouTube)", URL: "https://www.youtube.com/results?search_query=event+emcee+training+nepal"},
			}},
			{StepNumber: 6, Title: "Become a radio program director or launch your own show", Description: "Senior RJs can become program directors, managing the station's sound and schedule. Some launch their own independent podcasts or YouTube shows. Others start radio production companies. The skills of an RJ — voice, communication, creativity, audience understanding — are valuable in many fields. Radio might be your starting point, but a career in media has many directions. The key is to keep creating content, building your personal brand, and adapting to new platforms. Your voice and personality are your unique assets.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Radio program management (YouTube)", URL: "https://www.youtube.com/results?search_query=radio+program+management"},
				{Title: "Starting a podcast in Nepal", URL: "https://www.youtube.com/results?search_query=start+a+podcast+nepal"},
				{Title: "Media entrepreneurship Nepal", URL: "https://www.merojob.com"},
			}},
		},
	}
}

func tvPresenter() careerSeed {
	return careerSeed{
		CategoryName: "Media & Arts", CategorySlug: "media-arts", CategoryIcon: "📺",
		Title: "TV Presenter", Slug: "tv-presenter",
		Summary: "TV presenters host television programs, including news, talk shows, entertainment, sports, and special events.",
		Description: "A TV presenter appears on television to host programs, present news, interview guests, or entertain audiences. They are the face of TV channels and shows. In Nepal, TV presenters work for channels like Kantipur TV, Nepal Television, AP1 TV, Himalaya TV, News24, and more. Presenters need excellent communication skills, good appearance, confidence on camera, and the ability to handle live television. TV presenters can specialize in news reading, talk shows, sports commentary, entertainment shows, or live event coverage. The profession is glamorous but demanding — long hours, public scrutiny, and constant pressure to perform. Successful TV presenters become household names in Nepal.",
		DailyTasks: []string{"Research and prepare for shows — read scripts, study topics", "Rehearse segments and practice delivery", "Present live or recorded TV programs", "Interview guests — experts, celebrities, newsmakers", "Collaborate with producers, directors, and camera crew", "Attend promotional events and channel appearances", "Review recordings and improve performance"},
		Skills: []string{"Excellent verbal communication in Nepali and English", "Confidence and presence on camera", "Script reading and teleprompter skills", "Interviewing and conversation skills", "Research and preparation abilities", "Professional appearance and grooming", "Ability to handle live TV pressure", "Adaptability to different show formats"},
		SalaryMin: 360000, SalaryMax: 1800000, Difficulty: 3, FutureProof: 50,
		EducationReq: "Bachelor's in Mass Communication, Journalism, or related field. Voice and camera training from media institutes. Fluency in Nepali and English essential. Some presenters have backgrounds in modeling, theater, or public relations.",
		Outlook: "Nepal's television industry includes many private channels plus Nepal Television. Competition for on-air positions is intense. Digital video platforms (YouTube, TikTok) are creating new presenting opportunities. Many TV presenters also work as event hosts, brand ambassadors, and social media influencers. The skills are transferable across media.",
		Tags: []string{"media", "television", "communication", "entertainment", "hosting"},
		Resources: []resourceSeed{
			{Title: "Kantipur TV", URL: "https://www.kantipurtv.com", Description: "Major Nepali TV network and career opportunities"},
			{Title: "Nepal Television", URL: "https://www.ntv.com.np", Description: "National television broadcaster of Nepal"},
			{Title: "Merojob Media Jobs", URL: "https://www.merojob.com", Description: "Find TV presenting jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Develop your communication and presentation skills", Description: "TV presenting requires excellent spoken Nepali and English. Practice speaking clearly, confidently, and naturally. Record yourself on video and watch critically — improve your posture, gestures, facial expressions, and eye contact. Watch Nepali TV presenters and analyze what makes them effective. Notice how they use their voice, how they stand, how they transition between segments. Join a debate club, theater group, or public speaking forum to gain confidence. The camera magnifies every nervous habit — practice until you are comfortable being watched.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Public speaking and presentation skills (YouTube)", URL: "https://www.youtube.com/results?search_query=public+speaking+skills+in+nepali"},
				{Title: "Watch Nepali TV news presenters", URL: "https://www.kantipurtv.com"},
				{Title: "Body language tips for TV presenters", URL: "https://www.youtube.com/results?search_query=body+language+for+tv+presenters"},
			}},
			{StepNumber: 2, Title: "Get education in media and broadcasting", Description: "Enroll in a Mass Communication or Journalism degree program. Study television production, broadcast journalism, script writing, and media ethics. Learn how TV studios work — cameras, lights, sound, teleprompters, and production control. Participate in college TV projects or create your own YouTube channel to practice. A good education gives you the theoretical knowledge and practical experience that TV channels look for. Internships at TV stations during your studies are invaluable. They give you industry contacts and real studio experience.",
			Duration: "3-4 years", Links: []roadmapLink{
				{Title: "Mass Communication programs in Nepal (Edusanjal)", URL: "https://www.edusanjal.com"},
				{Title: "Television production basics (YouTube)", URL: "https://www.youtube.com/results?search_query=tv+production+basics+for+beginners"},
				{Title: "How a TV studio works", URL: "https://www.youtube.com/results?search_query=how+tv+studio+works+tour"},
			}},
			{StepNumber: 3, Title: "Start in smaller roles and build experience", Description: "Begin as a production assistant, researcher, or junior reporter for a TV channel. Learn the production process from behind the camera. Volunteer to do field reports or small segments that build your portfolio. Many successful TV presenters started as reporters. Reporting gives you credibility and experience being on camera in real situations. Create a showreel — a 2-3 minute video showing your best on-camera work. Your showreel is the most important tool for getting presenting jobs. Keep it updated and share it with TV channels and production houses.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Find media jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "How to make a TV presenter showreel (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+make+a+tv+presenter+showreel"},
				{Title: "Field reporting tips for TV", URL: "https://www.youtube.com/results?search_query=tv+field+reporting+tips+nepal"},
			}},
			{StepNumber: 4, Title: "Get your first on-air presenting role", Description: "Once you have experience and a good showreel, apply for on-air positions. Start with smaller shows — morning segments, weekend programs, or regional coverage. These roles have less pressure and let you develop your on-camera skills. Every show you present builds your confidence and visibility. Build relationships with producers and channel management. Be reliable, professional, and easy to work with. In television, your reputation is everything. The more shows you handle well, the better opportunities you will get.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "TV presenter career tips (YouTube)", URL: "https://www.youtube.com/results?search_query=tv+presenter+career+tips+nepal"},
				{Title: "How to handle live TV pressure", URL: "https://www.youtube.com/results?search_query=live+tv+presenting+tips"},
				{Title: "Nepal TV channel directory", URL: "https://www.youtube.com/results?search_query=nepali+tv+channels+list"},
			}},
			{StepNumber: 5, Title: "Specialize and build your personal brand", Description: "Develop expertise in a show format — news anchoring, talk shows, sports presenting, entertainment, or documentary hosting. Specialization makes you the go-to presenter for that format. Build your personal brand through social media. A TV presenter with a strong social media following is more valuable to channels. Attend media events and build your professional network. Many TV presenters also become brand ambassadors for companies. Your face and name are your brand — manage them carefully.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Personal branding for TV presenters (YouTube)", URL: "https://www.youtube.com/results?search_query=personal+branding+for+tv+presenters"},
				{Title: "Social media for TV personalities", URL: "https://www.youtube.com/results?search_query=social+media+tips+for+tv+hosts"},
				{Title: "News anchoring techniques", URL: "https://www.youtube.com/results?search_query=news+anchoring+techniques+nepal"},
			}},
			{StepNumber: 6, Title: "Become a senior presenter or media personality", Description: "Top presenters become the face of major shows or entire channels. They earn significantly more and have production input. Some move into media management, production, or start their own production companies. Others become full-time influencers, speakers, or consultants. A TV presenting career can lead to many other opportunities in media, entertainment, and public life. The key skills — communication, confidence, presence — are valuable everywhere. Successful TV presenters in Nepal are among the most recognized and influential people in the country.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "TV production company startup Nepal", URL: "https://www.merojob.com"},
				{Title: "From presenter to producer (YouTube)", URL: "https://www.youtube.com/results?search_query=from+tv+presenter+to+producer"},
				{Title: "Media career advancement in Nepal", URL: "https://www.youtube.com/results?search_query=media+career+growth+nepal"},
			}},
		},
	}
}

func photographer() careerSeed {
	return careerSeed{
		CategoryName: "Media & Arts", CategorySlug: "media-arts", CategoryIcon: "📷",
		Title: "Photographer", Slug: "photographer",
		Summary: "Photographers capture images for weddings, events, media, advertising, and art. They use cameras and editing software to create compelling visual content.",
		Description: "A photographer takes pictures for various purposes — weddings, portraits, news, advertising, fashion, products, or art. Photography in Nepal has grown significantly with the rise of social media and digital cameras. Nepali photographers work in diverse areas: wedding photography (a huge market in Nepal), photojournalism (covering news and social issues), landscape photography (Nepal's mountains and scenery), portrait photography, event photography, and commercial photography. Some specialize in specific areas like wildlife photography in Chitwan or Himalayan photography. Photography can be a rewarding career for creative people who have an eye for composition, light, and storytelling. Many photographers start with basic equipment and build their way up.",
		DailyTasks: []string{"Meet with clients to understand their photography needs", "Set up and adjust camera equipment for shoots", "Take photos at events, locations, or in studios", "Edit and retouch photos using Adobe Lightroom or Photoshop", "Organize and backup digital photo libraries", "Deliver final edited photos to clients", "Promote photography services on social media and websites"},
		Skills: []string{"Camera operation and knowledge of photography techniques", "Composition, lighting, and color theory", "Photo editing (Adobe Lightroom, Photoshop)", "Client communication and customer service", "Business management for freelance photographers", "Social media and portfolio website management", "Creativity and artistic vision", "Time management and reliability"},
		SalaryMin: 180000, SalaryMax: 1200000, Difficulty: 2, FutureProof: 50,
		EducationReq: "No formal degree required. Photography courses from institutes like Nepal College of Photography or Pixel Institute. Workshops and online tutorials (YouTube, Udemy). Apprenticeship with experienced photographers is the best learning path. Portfolio is more important than certificates.",
		Outlook: "Wedding photography remains the biggest market in Nepal. Social media has increased demand for quality photos. Smartphone photography has reduced demand for basic photography but increased demand for professional quality. Drone photography is a growing specialty. Competition is high — networking and portfolio quality determine success. Many photographers supplement income with videography.",
		Tags: []string{"media", "creative", "photography", "visual arts", "freelance"},
		Resources: []resourceSeed{
			{Title: "Nepal College of Photography", URL: "https://www.ncp.edu.np", Description: "Photography education and training in Nepal"},
			{Title: "Pixel Institute Nepal", URL: "https://www.pixelinstitute.com.np", Description: "Digital photography and filmmaking courses"},
			{Title: "Merojob Creative Jobs", URL: "https://www.merojob.com", Description: "Find photography jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Get a camera and start shooting every day", Description: "You do not need an expensive camera to start. A basic DSLR or even a good smartphone camera can work. Learn the fundamentals — aperture, shutter speed, ISO, and composition rules (rule of thirds, leading lines, symmetry). Take photos every day. Photograph everything — people, landscapes, objects, events. Practice is the only way to improve. Review your photos critically. What makes a good photo? Lighting, timing, composition, emotion. Join photography communities on Facebook or Instagram where Nepali photographers share work and give feedback. Photography is a skill that improves only through consistent practice and honest self-critique.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Photography basics for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=photography+basics+for+beginners+nepal"},
				{Title: "Nepali photographers community on Facebook", URL: "https://www.facebook.com"},
				{Title: "Understanding aperture, shutter, ISO tutorial", URL: "https://www.youtube.com/results?search_query=understanding+aperture+shutter+iso"},
			}},
			{StepNumber: 2, Title: "Learn photo editing and take a photography course", Description: "Photography is 50% taking photos and 50% editing. Learn Adobe Lightroom for color correction and Adobe Photoshop for advanced editing. Take a photography course at an institute in Nepal. A structured course teaches you things self-learning might miss — studio lighting, portrait techniques, business aspects. Learn about different photography genres and find what you enjoy most. Build a portfolio of your best work. Your portfolio is what clients will judge you by. Quality matters more than quantity. Show only your best 15-20 photos, not every photo you have taken.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Lightroom editing tutorial (YouTube)", URL: "https://www.youtube.com/results?search_query=lightroom+tutorial+for+beginners"},
				{Title: "Photoshop basics for photographers", URL: "https://www.youtube.com/results?search_query=photoshop+basics+for+photographers"},
				{Title: "Photography courses in Nepal (Pixel Institute)", URL: "https://www.pixelinstitute.com.np"},
			}},
			{StepNumber: 3, Title: "Start shooting for free or low cost to build portfolio", Description: "Offer free or low-cost photoshoots to friends, family, and community members. Volunteer to photograph events — weddings, festivals, school programs. These shoots build your portfolio and experience. The first 100 photoshoots are for learning, not for earning. Every shoot teaches you something — how to direct people, how to handle different lighting, how to be professional. Build a website or Instagram portfolio to showcase your work. Word of mouth is the most powerful marketing tool for photographers. Deliver excellent work to every client, even if they are paying nothing. They will recommend you to others.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Build a photography portfolio (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+build+a+photography+portfolio"},
				{Title: "Instagram portfolio tips for photographers", URL: "https://www.youtube.com/results?search_query=instagram+portfolio+tips+photographers"},
				{Title: "Wedding photography for beginners", URL: "https://www.youtube.com/results?search_query=wedding+photography+for+beginners+nepal"},
			}},
			{StepNumber: 4, Title: "Get professional equipment and specialize", Description: "As you earn from photography, invest in better equipment — full-frame camera, professional lenses (prime lenses for portraits, zoom for events), flash and lighting equipment. Specialize in a photography genre — wedding, portrait, product, landscape, photojournalism, or wildlife. Specialization allows you to charge higher rates and build a reputation. Wedding photography is the most lucrative market in Nepal. A good wedding photographer can earn 50,000-200,000 NPR per wedding. Build relationships with wedding planners, venues, and other vendors who can refer clients to you.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Wedding photography gear guide (YouTube)", URL: "https://www.youtube.com/results?search_query=wedding+photography+gear+guide"},
				{Title: "Nepal wedding photography market tips", URL: "https://www.youtube.com/results?search_query=nepal+wedding+photography+tips"},
				{Title: "Should you go full-frame? Camera buying guide", URL: "https://www.youtube.com/results?search_query=full+frame+vs+crop+sensor+photography"},
			}},
			{StepNumber: 5, Title: "Build your brand and charge professional rates", Description: "Create a professional brand — studio name, logo, website, pricing packages. Develop a signature style that clients recognize. Charge rates that reflect your skill level and experience. Professional rates for photography in Nepal range from 10,000 NPR for a portrait session to 200,000 NPR for a wedding. Build relationships with corporate clients — hotels, restaurants, real estate, advertising agencies. Corporate photography (products, events, marketing) provides consistent income. Teach photography workshops to supplement income and build your reputation. Publish your work in magazines or online platforms to gain recognition.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "How to price photography services (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+price+photography+services"},
				{Title: "Corporate photography business tips", URL: "https://www.youtube.com/results?search_query=corporate+photography+business"},
				{Title: "Building a photography brand in Nepal", URL: "https://www.pixelinstitute.com.np"},
			}},
			{StepNumber: 6, Title: "Expand into videography, studio ownership, or teaching", Description: "Many photographers expand into videography — wedding films, corporate videos, documentaries. Video is in high demand and pays well. Some open their own studio with permanent lighting setups and backdrops. Others become photography educators, teaching courses and conducting workshops. The best photographers are lifelong learners — photography technology evolves constantly, with new cameras, editing software, and techniques. Photography can provide a good living in Nepal if you combine artistic talent with business skills. Your photos capture moments that people treasure forever.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Transitioning from photography to videography (YouTube)", URL: "https://www.youtube.com/results?search_query=photography+to+videography+transition"},
				{Title: "Opening a photography studio in Nepal", URL: "https://www.merojob.com"},
				{Title: "Teaching photography tips", URL: "https://www.youtube.com/results?search_query=become+a+photography+teacher"},
			}},
		},
	}
}

func videoEditor() careerSeed {
	return careerSeed{
		CategoryName: "Media & Arts", CategorySlug: "media-arts", CategoryIcon: "🎬",
		Title: "Video Editor", Slug: "video-editor",
		Summary: "Video editors assemble raw footage into finished videos for films, TV, social media, advertisements, and events.",
		Description: "A video editor takes raw video footage and transforms it into a finished product. They cut, arrange, and enhance footage, add music and sound effects, insert graphics, and adjust colors. Video editing is a crucial part of filmmaking, TV production, YouTube content, advertising, and social media. Nepal's video production industry has grown with the rise of digital media — wedding films, music videos, corporate videos, documentaries, TV commercials, and YouTube channels all need editors. Good video editors combine technical skills (editing software proficiency) with creative skills (storytelling, pacing, visual taste). Video editing can be a freelance career or a permanent position at a production house. Remote work is common — many Nepali editors work for international clients online.",
		DailyTasks: []string{"Review raw footage and organize files for editing", "Cut and arrange clips to tell a story or convey a message", "Add transitions, effects, text, and graphics", "Mix audio — adjust levels, add music and sound effects", "Color correct and color grade footage for consistent look", "Export final video in appropriate formats for different platforms", "Collaborate with directors, producers, and clients on revisions"},
		Skills: []string{"Editing software expertise (Adobe Premiere Pro, DaVinci Resolve, Final Cut Pro)", "Storytelling and narrative sense", "Color correction and color grading", "Audio editing and sound design", "Motion graphics (After Effects basics)", "File organization and workflow management", "Client communication and revision handling", "Patience and attention to detail"},
		SalaryMin: 240000, SalaryMax: 1500000, Difficulty: 3, FutureProof: 70,
		EducationReq: "No formal degree required but preferred. Diploma or certificate in Video Editing from institutes like Pixel Institute, Nepal College of Film, or Broadway DSB. Online courses (LinkedIn Learning, Udemy, YouTube). Portfolio is more important than certificates.",
		Outlook: "Video content demand is exploding — YouTube, TikTok, Instagram Reels, and OTT platforms all need editors. Wedding films are a huge market in Nepal. Corporate video production is growing. Remote work opportunities for international clients are increasing. AI tools are changing editing workflows but creative editors remain valuable. Good editors are in high demand.",
		Tags: []string{"media", "creative", "video", "editing", "production"},
		Resources: []resourceSeed{
			{Title: "Pixel Institute Nepal", URL: "https://www.pixelinstitute.com.np", Description: "Film and video editing courses in Nepal"},
			{Title: "Broadway DSB Multimedia", URL: "https://www.broadwaydsb.com", Description: "Multimedia and video editing training"},
			{Title: "Adobe Premiere Pro tutorials (YouTube)", URL: "https://www.youtube.com/results?search_query=premiere+pro+tutorial+for+beginners+nepali"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn editing software and basic techniques", Description: "Start with free software like DaVinci Resolve (powerful and free) or learn Adobe Premiere Pro. Watch YouTube tutorials to learn the basics: cutting clips, arranging timelines, adding transitions, exporting videos. Practice by editing short videos using footage you shoot with your phone. Edit wedding footage, short films with friends, or social media clips. The only way to learn editing is by editing. Your first videos will be rough — that is normal. Keep practicing and learning new techniques. Editing requires patience. A 5-minute video can take 5 hours to edit properly.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "DaVinci Resolve beginner tutorial (YouTube)", URL: "https://www.youtube.com/results?search_query=davinci+resolve+beginner+tutorial+nepali"},
				{Title: "Premiere Pro for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=premiere+pro+beginner+tutorial"},
				{Title: "Video editing terminology explained", URL: "https://www.youtube.com/results?search_query=video+editing+terms+explained"},
			}},
			{StepNumber: 2, Title: "Take a professional editing course", Description: "Enroll in a video editing course at a recognized institute in Nepal. A structured course teaches you professional workflows, advanced techniques, and industry standards. You will learn color grading, audio mixing, motion graphics, and project management. Courses also provide access to professional software and feedback from experienced editors. Network with classmates and teachers — they are your future collaborators and clients. A good course can compress 2 years of self-learning into 3-6 months. Invest in your education — good editing skills earn good money.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Pixel Institute - Video Editing Course", URL: "https://www.pixelinstitute.com.np"},
				{Title: "Broadway DSB - Editing Programs", URL: "https://www.broadwaydsb.com"},
				{Title: "Color grading masterclass (YouTube)", URL: "https://www.youtube.com/results?search_query=color+grading+masterclass+davinci+resolve"},
			}},
			{StepNumber: 3, Title: "Build a portfolio with small projects", Description: "Edit videos for free or low cost to build your portfolio. Offer to edit wedding videos, YouTube videos for creators, music videos for local artists, or promotional videos for small businesses. Each project teaches you something new and adds to your showreel. Create a showreel — a 1-2 minute video showcasing your best editing work. Your showreel is your resume. Post your work on YouTube, Vimeo, and Instagram. Share before-and-after comparisons to demonstrate your skills. Client testimonials and word-of-mouth are powerful in Nepal's close-knit media industry.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "How to make a video editing showreel (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+make+a+video+editing+showreel"},
				{Title: "Freelance video editing for beginners", URL: "https://www.youtube.com/results?search_query=freelance+video+editing+for+beginners"},
				{Title: "Wedding video editing tips", URL: "https://www.youtube.com/results?search_query=wedding+video+editing+tips+nepal"},
			}},
			{StepNumber: 4, Title: "Specialize in a type of video editing", Description: "Develop expertise in one or more editing specialties: wedding films (huge market in Nepal), music videos (creative editing), corporate videos (clean professional style), documentary editing (long-form storytelling), social media content (fast-paced short format), or motion graphics (After Effects animation). Specialization makes you more valuable and allows you to charge higher rates. Learn After Effects for motion graphics and visual effects. Motion graphics editors are in high demand for advertising and explainer videos. The more specialized your skills, the less competition you face.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Wedding film editing style guide (YouTube)", URL: "https://www.youtube.com/results?search_query=wedding+film+editing+style+nepal"},
				{Title: "After Effects for beginners", URL: "https://www.youtube.com/results?search_query=after+effects+beginner+tutorial"},
				{Title: "Documentary editing techniques", URL: "https://www.youtube.com/results?search_query=documentary+editing+techniques"},
			}},
			{StepNumber: 5, Title: "Build a client base and professional network", Description: "Start charging professional rates. Build relationships with production houses, advertising agencies, YouTube channels, and event management companies in Nepal. Create a professional website or portfolio page. Use social media to showcase your work. Join Nepali filmmakers and editors groups on Facebook. Networking is essential — many editing jobs come through referrals. Develop good client communication skills. Reliable editors who deliver on time and handle revisions professionally get repeat business and referrals. Consider working on freelance platforms (Upwork, Fiverr) for international clients who pay in foreign currency.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Freelance video editing on Upwork (YouTube)", URL: "https://www.youtube.com/results?search_query=upwork+video+editing+guide+nepal"},
				{Title: "How to price video editing services", URL: "https://www.youtube.com/results?search_query=how+to+price+video+editing+services"},
				{Title: "Client communication tips for editors (YouTube)", URL: "https://www.youtube.com/results?search_query=client+communication+for+video+editors"},
			}},
			{StepNumber: 6, Title: "Become a lead editor or start your own production company", Description: "Senior editors become lead editors or post-production supervisors, managing teams of editors and the entire post-production workflow. Some start their own video production companies. Starting a production company requires business skills, equipment investment, and client relationships. The demand for video content is growing faster than the supply of good editors. A skilled editor with good business sense can build a successful career in Nepal. Every video you edit is watched by people who may become your next client. Quality work markets itself.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a video production company in Nepal", URL: "https://www.merojob.com"},
				{Title: "Post-production workflow management (YouTube)", URL: "https://www.youtube.com/results?search_query=post+production+workflow+management"},
				{Title: "Video production business tips", URL: "https://www.youtube.com/results?search_query=video+production+business+nepal"},
			}},
		},
	}
}

func contentWriter() careerSeed {
	return careerSeed{
		CategoryName: "Media & Arts", CategorySlug: "media-arts", CategoryIcon: "✍️",
		Title: "Content Writer", Slug: "content-writer",
		Summary: "Content writers create written content for websites, blogs, social media, marketing materials, and more. They write words that inform, persuade, and engage readers.",
		Description: "A content writer produces written material for digital and print media. This includes blog posts, website content, social media posts, email newsletters, product descriptions, white papers, case studies, and advertising copy. Content writing has grown enormously with digital marketing. Every business needs content for their website and social media. Nepali content writers work for agencies, companies, or as freelancers. Good writers are in high demand. Content writing is one of the most accessible remote work careers — you can work from anywhere with an internet connection and earn in foreign currency. Many Nepali freelancers work for international clients through platforms like Upwork, Fiverr, and ProBlogger.",
		DailyTasks: []string{"Research topics to understand subject matter and find reliable sources", "Write blog posts, articles, web pages, or social media content", "Edit and proofread content for grammar, clarity, and accuracy", "Optimize content for search engines (SEO) with keywords", "Collaborate with clients or marketing teams on content strategy", "Meet deadlines and manage multiple writing projects", "Stay updated on industry trends and writing best practices"},
		Skills: []string{"Excellent writing and grammar in English and/or Nepali", "Research and fact-checking ability", "SEO (search engine optimization) knowledge", "Understanding of different content formats and styles", "Time management and meeting deadlines", "Basic WordPress or CMS skills", "Ability to accept and apply feedback", "Creativity and adaptability in writing style"},
		SalaryMin: 180000, SalaryMax: 900000, Difficulty: 2, FutureProof: 60,
		EducationReq: "Bachelor's degree in English, Journalism, Mass Communication, or related field preferred but not required. Strong writing skills are more important than degrees. Online courses in content writing, SEO writing, and copywriting from platforms like Coursera, HubSpot Academy, and Google Digital Garage.",
		Outlook: "Content writing demand is growing with digital marketing. Businesses need content for websites, blogs, and social media. Remote work opportunities for international clients are abundant. AI writing tools are changing the field but human writers are still needed for quality, creativity, and strategy. Specialized writers (technical, medical, legal) earn more.",
		Tags: []string{"media", "writing", "content", "digital marketing", "freelance"},
		Resources: []resourceSeed{
			{Title: "Merojob Content Jobs", URL: "https://www.merojob.com", Description: "Find content writing jobs in Nepal"},
			{Title: "Google Digital Garage Nepal", URL: "https://www.digitalunlocked.com/np", Description: "Free digital marketing and content courses"},
			{Title: "HubSpot Content Marketing Certification", URL: "https://academy.hubspot.com", Description: "Free content marketing certification"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Read and write every day to develop your craft", Description: "Good writers are good readers first. Read books, articles, blogs, and newspapers daily. Notice how different writers structure their work, use language, and engage readers. Write every day — start a blog, write social media posts, keep a journal, write articles on topics you care about. Quantity leads to quality. Your first 100 articles will not be great. That is fine. Keep writing. Join writing communities online where you can share your work and get feedback. Develop a thick skin—criticism makes you better. Writing is a craft that improves only through consistent practice.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Blogging tips for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=blogging+tips+for+beginners+nepal"},
				{Title: "How to start a blog in Nepal", URL: "https://www.youtube.com/results?search_query=start+a+blog+nepal"},
				{Title: "Nepali writing community on Facebook", URL: "https://www.facebook.com"},
			}},
			{StepNumber: 2, Title: "Learn SEO content writing and digital marketing basics", Description: "Content writing for the web is different from academic writing. Learn SEO (Search Engine Optimization) — how to use keywords, write meta descriptions, structure content for search engines, and optimize for readers. Understand digital marketing basics — how content fits into marketing strategy. Take free courses from HubSpot Academy, Google Digital Garage, or Semrush Academy. Learn about different content types: blog posts, landing pages, email newsletters, social media content, product descriptions. Each type has different requirements and styles. Knowing SEO makes you significantly more valuable as a content writer.",
			Duration: "2-4 months", Links: []roadmapLink{
				{Title: "SEO content writing for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=seo+content+writing+for+beginners"},
				{Title: "HubSpot Content Marketing Certification", URL: "https://academy.hubspot.com"},
				{Title: "Google Digital Garage Nepal", URL: "https://www.digitalunlocked.com/np"},
			}},
			{StepNumber: 3, Title: "Build a portfolio with samples", Description: "Create writing samples in different formats — blog posts, web pages, social media posts, email newsletters. If you have no clients, write for imaginary businesses or volunteer to write for local organizations. Start a blog on a topic you are passionate about. Your blog serves as both practice and portfolio. Guest post on other websites to build your portfolio and gain exposure. Create a simple portfolio website or use platforms like Medium or LinkedIn to showcase your writing. Your writing samples are what clients judge you by. Make sure your best work is easy to find and read.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "How to build a writing portfolio (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+build+a+writing+portfolio"},
				{Title: "Guest blogging tips for exposure", URL: "https://www.youtube.com/results?search_query=guest+blogging+tips"},
				{Title: "Creating a portfolio website with WordPress", URL: "https://www.youtube.com/results?search_query=wordpress+portfolio+website+nepal"},
			}},
			{StepNumber: 4, Title: "Start freelancing on platforms", Description: "Create profiles on freelance platforms like Upwork, Fiverr, and Freelancer. Start with lower rates to build reviews and reputation. Apply to many jobs — expect rejection, but keep applying. As you build positive reviews and a work history, raise your rates. Many Nepali content writers earn $500-$3000 per month working for international clients. Learn to write compelling proposals that show you understand the client's needs. Deliver high-quality work on time. Happy clients give good reviews and repeat work. Freelancing teaches you client management, time management, and business skills.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Upwork guide for Nepali freelancers (YouTube)", URL: "https://www.youtube.com/results?search_query=upwork+for+nepali+freelancers"},
				{Title: "How to write winning proposals on Fiverr", URL: "https://www.youtube.com/results?search_query=fiverr+winning+proposal+tips"},
				{Title: "Freelance content writing rates (YouTube)", URL: "https://www.youtube.com/results?search_query=content+writing+rates+for+beginners"},
			}},
			{StepNumber: 5, Title: "Specialize in a content niche", Description: "Specialized content writers earn more than general writers. Popular niches include: technology writing, finance and crypto, health and medical, travel, lifestyle, real estate, and legal writing. Choose a niche that interests you and learn deeply about it. Specialized knowledge combined with excellent writing skills makes you a premium writer. You can charge 2-5x more than general writers. Build expertise by reading industry publications, following thought leaders, and creating niche content samples. A specialist writer is harder to replace with AI than a general writer.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Niche content writing strategies (YouTube)", URL: "https://www.youtube.com/results?search_query=niche+content+writing+strategy"},
				{Title: "Technical writing career guide", URL: "https://www.youtube.com/results?search_query=technical+writing+career+nepal"},
				{Title: "Travel writing for Nepal destinations", URL: "https://www.youtube.com/results?search_query=travel+writing+nepal"},
			}},
			{StepNumber: 6, Title: "Become a content strategist or editor", Description: "Experienced content writers become content strategists — planning entire content calendars, managing content teams, and aligning content with business goals. Others become editors, overseeing quality and guiding junior writers. Some start their own content agencies serving multiple clients. Content strategy pays significantly more than writing alone. Build skills in analytics to measure content performance, content management systems, and content marketing strategy. The most valuable content professionals combine writing skills with strategic thinking. Content is how businesses communicate with the world — good content is always in demand.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Content strategy career path (YouTube)", URL: "https://www.youtube.com/results?search_query=content+strategy+career+path"},
				{Title: "Starting a content writing agency", URL: "https://www.youtube.com/results?search_query=start+a+content+agency+nepal"},
				{Title: "Editorial career progression guide", URL: "https://www.youtube.com/results?search_query=editor+career+path"},
			}},
		},
	}
}

func painter() careerSeed {
	return careerSeed{
		CategoryName: "Media & Arts", CategorySlug: "media-arts", CategoryIcon: "🎨",
		Title: "Painter (Artist)", Slug: "painter-artist",
		Summary: "Painters create original artworks using various mediums—oil, acrylic, watercolor, or digital. They sell their work through galleries, commissions, and online.",
		Description: "A painter creates visual art using various materials and techniques. Nepal has a rich artistic tradition with famous artists like Lain Singh Bangdel, Utpala Shrestha, and Kiran Manandhar. Painters work in oil, acrylic, watercolor, pastel, or digital media. Some focus on traditional Nepali styles — paubha (thangka) painting, Mithila art, or contemporary Nepali art. Others work in international styles. Painters earn through gallery sales, commissions (portraits, murals, corporate art), teaching art, illustrating books, and increasingly through online platforms. The art market in Nepal is growing but remains challenging. Successful painters combine artistic talent with business and marketing skills. Art is both a passion and a business.",
		DailyTasks: []string{"Plan and sketch new artwork concepts", "Paint — mix colors, apply paint, develop compositions", "Prepare canvases and art materials", "Photograph and document completed works for portfolio", "Promote art on social media (Instagram, Facebook)", "Communicate with galleries, clients, and collectors", "Participate in exhibitions and art events"},
		Skills: []string{"Drawing and painting techniques in chosen medium", "Color theory and composition", "Creativity and personal artistic vision", "Knowledge of art history and contemporary trends", "Photography to document artwork professionally", "Social media marketing and self-promotion", "Client communication for commissions", "Business management for artists"},
		SalaryMin: 120000, SalaryMax: 1200000, Difficulty: 3, FutureProof: 40,
		EducationReq: "Bachelor's in Fine Arts (BFA) from TU, Kathmandu University, or Lalitkala Campus. Diploma in Fine Arts. Formal training is helpful but not essential — many successful artists are self-taught. Continuous practice and portfolio development is what matters most.",
		Outlook: "Nepal's art scene is small but active with galleries in Kathmandu and Pokhara. Online sales through Instagram and Etsy are growing. Art tourism (selling to visitors) is significant. Government and corporate art purchases are limited. Many artists supplement income through teaching or commercial art. Digital art and NFTs are creating new opportunities.",
		Tags: []string{"media", "arts", "creative", "traditional", "digital art"},
		Resources: []resourceSeed{
			{Title: "Nepal Art Council", URL: "https://www.nepalartcouncil.com", Description: "Nepal's premier art institution and gallery"},
			{Title: "Lalitkala Campus (TU)", URL: "https://lalitkalacampus.edu.np", Description: "Fine arts education in Kathmandu"},
			{Title: "Siddhartha Art Gallery", URL: "https://www.siddharthaartgallery.com", Description: "Contemporary art gallery in Kathmandu"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Develop your artistic skills through daily practice", Description: "Art is 1% talent and 99% practice. Draw and paint every single day. Experiment with different mediums — pencil, charcoal, watercolor, acrylic, oil. Find what medium suits your style and expression. Study the fundamentals: perspective, proportion, color theory, composition, light and shadow. Copy the masters to learn techniques. Visit art galleries in Kathmandu (Siddhartha, Nepal Art Council, Taragaon Museum) to see the work of established artists. Your skill will improve only through consistent practice. Keep all your work — looking back at earlier pieces shows your progress.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Drawing basics for absolute beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=drawing+basics+for+beginners"},
				{Title: "Color theory for painters", URL: "https://www.youtube.com/results?search_query=color+theory+for+painters+tutorial"},
				{Title: "Nepali painting styles and traditions", URL: "https://www.youtube.com/results?search_query=nepali+painting+styles+traditional"},
			}},
			{StepNumber: 2, Title: "Get formal art education or mentorship", Description: "Enroll in a Fine Arts program at TU's Lalitkala Campus, Kathmandu University School of Arts, or other art institutes. A formal education provides structure, exposure to art history, feedback from professors, and a community of artists. Alternatively, find a mentor — an established artist who can guide you. Many Nepali artists offer studio apprenticeships or workshops. Art education teaches you not just technique but also how to think about art, develop concepts, and present your work professionally. The connections you make in art school are valuable for your entire career.",
			Duration: "3-4 years", Links: []roadmapLink{
				{Title: "Lalitkala Campus - BFA Program", URL: "https://lalitkalacampus.edu.np"},
				{Title: "Kathmandu University School of Arts", URL: "https://www.ku.edu.np"},
				{Title: "Art workshops in Kathmandu", URL: "https://www.youtube.com/results?search_query=art+workshops+nepal+kathmandu"},
			}},
			{StepNumber: 3, Title: "Create a body of work and build your portfolio", Description: "Develop a consistent body of work — a collection of paintings that show your style and themes. A strong artist has a recognizable voice that makes their work identifiable. Create at least 20-30 finished pieces for your portfolio. Photograph your work professionally — good photos are essential for your portfolio, social media, and submissions. Write artist statements and bios. Your portfolio is how galleries and buyers judge your work. Quality and consistency matter more than quantity. Your body of work tells the story of who you are as an artist.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "How to photograph your artwork (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+photograph+artwork+professionally"},
				{Title: "Writing an artist statement guide", URL: "https://www.youtube.com/results?search_query=how+to+write+an+artist+statement"},
				{Title: "Building an art portfolio for beginners", URL: "https://www.youtube.com/results?search_query=build+art+portfolio+tips"},
			}},
			{StepNumber: 4, Title: "Exhibit your work and build your audience", Description: "Start by participating in group exhibitions at local galleries. Submit your work to open calls and art competitions. Have your first solo exhibition when you have a strong body of work. Use social media (Instagram is essential for visual artists) to share your work daily. Build an audience who appreciates your art. Attend gallery openings and art events to network. The art world runs on relationships. Meet gallery owners, curators, collectors, and other artists. Be professional in all your dealings. An artist who is reliable and easy to work with gets more opportunities.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Nepal Art Council - Exhibition opportunities", URL: "https://www.nepalartcouncil.com"},
				{Title: "Instagram marketing for artists (YouTube)", URL: "https://www.youtube.com/results?search_query=instagram+marketing+for+artists"},
				{Title: "How to get your first gallery exhibition", URL: "https://www.youtube.com/results?search_query=how+to+get+a+gallery+exhibition"},
			}},
			{StepNumber: 5, Title: "Sell work and build your art business", Description: "Develop multiple income streams: gallery sales, commissions, teaching art classes, selling prints, book illustration, mural commissions, and online sales. Set professional pricing for your work. Understand the market in Nepal — prices range from 5,000 NPR for emerging artists to 500,000+ NPR for established names. Build relationships with art buyers and collectors. Nepal's art market is small but passionate collectors support local artists. Consider selling online through Instagram, Etsy, or Saatchi Art. International sales can be significant if you build a global audience. Art is both inspiration and business.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Pricing your art as an emerging artist (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+price+your+artwork+tips"},
				{Title: "Selling art online - Etsy guide", URL: "https://www.youtube.com/results?search_query=selling+art+on+etsy+tips"},
				{Title: "Art teaching as a career option", URL: "https://www.youtube.com/results?search_query=become+an+art+teacher+nepal"},
			}},
			{StepNumber: 6, Title: "Establish yourself as a recognized artist", Description: "Established artists in Nepal are those with consistent exhibition history, media coverage, collector base, and recognition. They may receive public art commissions, international residencies, and teaching positions at universities. Art is a long game — most artists take 10-20 years to reach their full potential. Stay true to your artistic vision while adapting to market realities. The most successful artists in Nepal are those who persisted through challenges and continued creating. Your art is your contribution to culture. Paintings outlive their creators — you are creating something that may inspire people for generations.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Art residencies and grants for Nepali artists", URL: "https://www.youtube.com/results?search_query=art+residencies+for+nepali+artists"},
				{Title: "Public art commissions in Nepal", URL: "https://www.nepalartcouncil.com"},
				{Title: "Building a lasting career as an artist (YouTube)", URL: "https://www.youtube.com/results?search_query=long+term+art+career+strategy"},
			}},
		},
	}
}

func musician() careerSeed {
	return careerSeed{
		CategoryName: "Media & Arts", CategorySlug: "media-arts", CategoryIcon: "🎵",
		Title: "Musician", Slug: "musician",
		Summary: "Musicians create, perform, and record music. They play instruments, sing, compose songs, and entertain audiences in Nepal and beyond.",
		Description: "A musician creates and performs music. Nepal has a vibrant music scene spanning folk, classical, pop, rock, hip-hop, and devotional music. Musicians can be vocalists, instrumentalists (guitar, drums, sarangi, madal, flute), songwriters, composers, or music producers. The Nepali music industry includes live performances (concerts, events, restaurant gigs), recording studios, music videos, film music, and increasingly digital distribution through YouTube and streaming platforms. Many Nepali musicians combine performance with teaching music or working in recording studios. The music industry is competitive but rewarding for talented and persistent musicians. Success requires not just musical talent but also self-promotion, networking, and business skills.",
		DailyTasks: []string{"Practice instrument or voice — scales, techniques, songs", "Write and compose new songs or instrumental pieces", "Rehearse with band or group for upcoming performances", "Record music in studio or home recording setup", "Perform live at concerts, events, or venues", "Promote music on social media, YouTube, and streaming platforms", "Network with other musicians, producers, and industry contacts"},
		Skills: []string{"Instrumental or vocal proficiency in chosen style", "Music theory knowledge (notes, chords, rhythm, harmony)", "Songwriting and composition ability", "Performance skills and stage presence", "Recording and music production basics", "Social media and YouTube content creation", "Collaboration and teamwork with other musicians", "Persistence and resilience in a competitive field"},
		SalaryMin: 120000, SalaryMax: 1500000, Difficulty: 3, FutureProof: 40,
		EducationReq: "Formal music education from Nepal Music Academy, Thyagaraja Music College, or Prajna Music College. Many musicians are self-taught. Mentorship from established musicians is invaluable. Online learning (YouTube, MasterClass). Theory and notation knowledge helps but passion and practice matter more.",
		Outlook: "Nepali music industry is growing with digital platforms. YouTube and streaming provide new income and exposure. Live performances and event gigs are recovering post-pandemic. Film music remains a major opportunity. Competition is high. Many musicians need multiple income streams (teaching, gigging, studio work, session playing). The industry is tough but passionate musicians find ways.",
		Tags: []string{"media", "music", "entertainment", "creative", "performance"},
		Resources: []resourceSeed{
			{Title: "Nepal Music Academy", URL: "https://www.nepalmusicacademy.com", Description: "Music education and training in Nepal"},
			{Title: "Music Nepal - Recording Studio", URL: "https://www.musicnepal.com", Description: "Major Nepali music production and distribution company"},
			{Title: "Nepali music YouTube channels", URL: "https://www.youtube.com", Description: "Nepali songs and music content on YouTube"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn your instrument or voice with daily practice", Description: "Music mastery comes from consistent practice — hours every day. Learn the fundamentals of your instrument or voice. Take lessons from a teacher if possible. Learn music theory — notes, scales, chords, rhythm, harmony. Theory is the language of music and helps you communicate with other musicians. Learn to read music notation and tablature. Practice both technical exercises and actual songs. Record yourself playing and listen critically. Find your strengths and weaknesses. Join a band or group as early as possible — playing with others is the best learning experience. Music is a lifelong journey of learning and improvement.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Guitar lessons for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=guitar+lessons+for+beginners+nepali"},
				{Title: "Vocal training exercises for beginners", URL: "https://www.youtube.com/results?search_query=vocal+training+for+beginners"},
				{Title: "Music theory basics explained simply", URL: "https://www.youtube.com/results?search_query=music+theory+for+beginners+explained"},
			}},
			{StepNumber: 2, Title: "Get formal music education or mentorship", Description: "Enroll in a music college or join a reputable music school in Nepal. Study music theory, history, composition, and performance. Formal education provides structure, exposure to different genres, and access to experienced teachers. Many Nepali musicians also learn through the guru-shishya (teacher-student) tradition — finding an experienced musician to mentor you. Learn about the Nepali music industry — how songs are produced, distributed, and promoted. Study both Nepali folk/classical traditions and Western music. The best musicians are versatile and understand multiple styles.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Thyagaraja Music College Nepal", URL: "https://www.thyagarajamusiccollege.com"},
				{Title: "Prajna Music College", URL: "https://www.prajnamusiccollege.org"},
				{Title: "Nepali folk music traditions course", URL: "https://www.youtube.com/results?search_query=nepali+folk+music+lessons"},
			}},
			{StepNumber: 3, Title: "Perform live and build your presence", Description: "Start performing wherever you can — open mic nights, college events, restaurants, hotels, street performances. Live performance is where you develop stage presence, learn to handle audiences, and gain confidence. Record your performances and share on social media. Build a following on YouTube — post covers, original songs, and music videos. YouTube is the most important platform for Nepali musicians today. Collaborate with other musicians — each collaboration introduces you to their audience. Be persistent. The Nepali music industry is small and relationships matter. Every person you meet could be a future collaborator.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Open mic nights in Kathmandu", URL: "https://www.facebook.com"},
				{Title: "Growing your YouTube music channel (YouTube)", URL: "https://www.youtube.com/results?search_query=grow+youtube+music+channel+nepal"},
				{Title: "Live performance tips for musicians", URL: "https://www.youtube.com/results?search_query=live+performance+tips+for+musicians"},
			}},
			{StepNumber: 4, Title: "Record and release your music professionally", Description: "Record your songs in a professional studio in Nepal (Music Nepal, Kalinchok Studio, OM Recording Studio). Work with experienced producers and sound engineers. Release your music on streaming platforms (Spotify, Apple Music) through a distributor (DistroKid, TuneCore, RouteNote). Create music videos for YouTube. Nepali music videos are popular and can generate millions of views. Develop your unique sound and style. The Nepali music scene is diverse — from folk-rock to pop to hip-hop. Find your niche and audience. Release consistently — one song every few months keeps your audience engaged and growing.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Recording studios in Kathmandu (YouTube)", URL: "https://www.youtube.com/results?search_query=recording+studio+kathmandu+nepal"},
				{Title: "How to release music on Spotify (YouTube)", URL: "https://www.youtube.com/results?search_query=release+music+on+spotify+guide"},
				{Title: "Nepali music video production tips", URL: "https://www.youtube.com/results?search_query=nepali+music+video+making+tips"},
			}},
			{StepNumber: 5, Title: "Diversify income streams as a musician", Description: "Few musicians earn enough from music sales and streaming alone. Build multiple income streams: live performances (concerts, events, restaurant/hotel gigs), teaching music (private lessons, group classes, workshops), session work (playing on other artists' recordings), composition and arrangement for film/TV/advertising, selling merchandise, YouTube monetization, and music production for other artists. Music teaching is a stable income source for many Nepali musicians. Understand the business side — contracts, royalties, copyright, and licensing. A musician who understands business is more likely to succeed.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Music teaching career in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=music+teaching+career+nepal"},
				{Title: "Music copyright and royalties explained", URL: "https://www.youtube.com/results?search_query=music+copyright+royalties+explained"},
				{Title: "Session musician career guide", URL: "https://www.youtube.com/results?search_query=become+session+musician+nepal"},
			}},
			{StepNumber: 6, Title: "Become an established name in Nepali music", Description: "Top Nepali musicians are those who have built a loyal fan base, have multiple hit songs, perform at major venues, and influence the music scene. They may work on film music (a major opportunity in Nepal's film industry), produce other artists, or run their own music label. Music is a marathon, not a sprint. Most successful musicians took years or decades to establish themselves. Stay true to your artistic vision while adapting to changing trends. Your music can touch people's hearts, express emotions they cannot express, and become part of Nepal's cultural fabric.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Nepali film music industry opportunities", URL: "https://www.youtube.com/results?search_query=nepali+film+music+career"},
				{Title: "Starting a music label in Nepal", URL: "https://www.merojob.com"},
				{Title: "Building a long-term music career (YouTube)", URL: "https://www.youtube.com/results?search_query=build+sustainable+music+career"},
			}},
		},
	}
}
func civilServiceOfficer() careerSeed {
	return careerSeed{
		CategoryName: "Government & Law", CategorySlug: "government-law", CategoryIcon: "🏛️",
		Title: "Civil Service Officer (Loksewa)", Slug: "civil-service-officer",
		Summary: "Civil service officers implement government policies, deliver public services, and manage government departments at federal, provincial, and local levels in Nepal.",
		Description: "A civil service officer (Loksewa officer) works for the Government of Nepal to implement policies, deliver services to citizens, and manage government administration. Civil servants are recruited through the Public Service Commission (Loksewa Ayog) exams. Nepal's civil service has multiple levels — gazetted officers (first, second, third class) and non-gazetted staff. Officers work in various ministries, departments, and local governments. Civil service offers job security, benefits, and the opportunity to serve the nation. Positions range from general administration to specialized technical roles (engineering, health, agriculture, education). The Loksewa exam is one of the most competitive exams in Nepal. Preparation requires dedication and discipline.",
		DailyTasks: []string{"Review files, correspondence, and policy documents", "Meet with citizens to address complaints and provide services", "Draft reports, memos, and official correspondence", "Coordinate with other departments and levels of government", "Supervise subordinate staff and monitor work progress", "Attend meetings with officials, stakeholders, and committees", "Implement and monitor government programs and projects"},
		Skills: []string{"Knowledge of Nepali laws, rules, and government procedures", "Strong writing and documentation skills in Nepali and English", "Analytical and problem-solving ability", "Leadership and personnel management", "Public relations and citizen service orientation", "Computer skills for office work", "Integrity, honesty, and commitment to public service", "Ability to work within bureaucratic systems"},
		SalaryMin: 350000, SalaryMax: 2000000, Difficulty: 4, FutureProof: 75,
		EducationReq: "Bachelor's degree minimum for officer-level positions. Master's preferred for senior positions. Loksewa exam preparation through coaching institutes (Unique Academy, Career Makers, IQRA Academy). Nepal Civil Service Act determines qualifications for different positions.",
		Outlook: "Civil service remains one of the most desirable careers in Nepal due to job security, benefits, and pension. Competition is intense with thousands of applicants for few positions. The federal structure has created more positions at provincial and local levels. Retirements will create openings. Civil service reform is ongoing.",
		Tags: []string{"government", "administration", "public service", "stable"},
		Resources: []resourceSeed{
			{Title: "Public Service Commission Nepal (Loksewa)", URL: "https://www.psc.gov.np", Description: "Official website for Loksewa exams and recruitment"},
			{Title: "Merojob Government Jobs", URL: "https://www.merojob.com", Description: "Find government job openings in Nepal"},
			{Title: "Loksewa preparation materials", URL: "https://www.youtube.com/results?search_query=loksewa+preparation+guide+nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Excel in your bachelor's degree and general knowledge", Description: "A strong academic foundation is essential for Loksewa preparation. Focus on your bachelor's degree in any field — but subjects like public administration, law, political science, economics, and sociology are particularly relevant. Read Nepali and English newspapers daily (Gorkhapatra, Kathmandu Post, Kantipur). Build general knowledge about Nepal's history, geography, constitution, current affairs, and international relations. The Loksewa exam tests general knowledge extensively. Subscribe to Loksewa preparation YouTube channels. The journey to becoming a civil servant starts with building a strong foundation of knowledge.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "PSC Nepal - Exam Information", URL: "https://www.psc.gov.np"},
				{Title: "Daily news for Loksewa preparation (YouTube)", URL: "https://www.youtube.com/results?search_query=daily+current+affairs+for+loksewa"},
				{Title: "Nepal's constitution explained (YouTube)", URL: "https://www.youtube.com/results?search_query=nepal+constitution+explained"},
			}},
			{StepNumber: 2, Title: "Join a Loksewa coaching institute", Description: "Enroll in a Loksewa coaching institute in Kathmandu or your province. Institutes like Unique Academy, IQRA Academy, and Career Makers have structured programs for the Loksewa exam. Coaching provides study materials, mock tests, experienced teachers, and a competitive environment. The Loksewa exam has multiple stages: written exam (general paper + optional subject), interview, and sometimes practical tests. Learn the exam pattern, marking scheme, and time management strategies. Coaching institutes also help with optional subject selection — choose a subject you are strong in. Dedicated preparation for 6-12 months is typically needed for the officer-level exam.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Unique Academy - Loksewa Preparation", URL: "https://www.youtube.com/results?search_query=unique+academy+loksewa+nepal"},
				{Title: "IQRA Academy Loksewa Classes", URL: "https://www.youtube.com/results?search_query=iqra+academy+loksewa"},
				{Title: "Loksewa exam pattern and strategy (YouTube)", URL: "https://www.youtube.com/results?search_query=loksewa+exam+pattern+and+strategy"},
			}},
			{StepNumber: 3, Title: "Pass the Loksewa written exam", Description: "The written exam is the first and most competitive stage. General paper covers Nepali, English, general knowledge, and service-related topics. The optional subject paper tests your chosen specialization. Prepare systematically — daily study schedule, regular revision, and mock tests. Write practice answers to develop speed and presentation. Answer writing skill is crucial for Loksewa exams. Join study groups for motivation and shared learning. The written exam typically has a pass rate of 1-5%. Do not be discouraged by failure — many successful officers passed on their second or third attempt. Persistence is key.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Loksewa written exam tips (YouTube)", URL: "https://www.youtube.com/results?search_query=loksewa+written+exam+tips+nepal"},
				{Title: "Answer writing skills for Loksewa", URL: "https://www.youtube.com/results?search_query=answer+writing+for+loksewa+exam"},
				{Title: "Loksewa previous year question papers", URL: "https://www.psc.gov.np"},
			}},
			{StepNumber: 4, Title: "Prepare for and pass the interview", Description: "After passing the written exam, you face the interview. Interviews test your personality, communication skills, general awareness, and suitability for civil service. Be prepared with current affairs, your educational background, and your motivation for civil service. Practice mock interviews. The interview panel includes senior civil servants and subject experts. Confidence, clarity, and honesty matter more than knowing everything. If you do not know something, say so honestly rather than bluffing. The interview typically counts 10-20% of the final score. Good interview performance can significantly improve your ranking.",
			Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Loksewa interview tips and questions (YouTube)", URL: "https://www.youtube.com/results?search_query=loksewa+interview+questions+and+answers"},
				{Title: "Mock interview for civil service", URL: "https://www.youtube.com/results?search_query=civil+service+mock+interview+nepal"},
				{Title: "Common Loksewa interview mistakes", URL: "https://www.youtube.com/results?search_query=loksewa+interview+common+mistakes"},
			}},
			{StepNumber: 5, Title: "Complete training and start your civil service career", Description: "After selection, you attend training at the Nepal Administrative Training Academy (formerly Civil Service Training Academy) in Jawalakhel. Training covers government procedures, financial management, public service delivery, and leadership. Then you receive your posting — you could be assigned to any district in Nepal. Be open to serving anywhere. The early years of civil service involve learning the practical aspects of government work. Build relationships with colleagues, understand the local context, and deliver quality service to citizens. Your reputation as a capable and honest officer will determine your career progression.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Nepal Administrative Training Academy", URL: "https://www.nata.gov.np"},
				{Title: "Life of a civil service officer in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=life+of+a+loksewa+officer+nepal"},
				{Title: "Government financial rules and procedures Nepal", URL: "https://www.youtube.com/results?search_query=nepal+government+financial+procedures"},
			}},
			{StepNumber: 6, Title: "Advance through promotions and specialized training", Description: "Civil service careers progress through promotions based on performance, seniority, and additional exams. Pursue in-service training, foreign study opportunities, and specialized certifications. Senior positions include under-secretary, joint-secretary, and secretary of ministries. Some officers become chief of provincial or local government units. The top civil service position is Chief Secretary of the Government of Nepal. Throughout your career, maintain integrity and dedication to public service. Civil service is not just a job — it is a responsibility to serve Nepal and its people. Good civil servants build the nation.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Civil service career progression Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=civil+service+promotion+nepal"},
				{Title: "In-service training for Nepali civil servants", URL: "https://www.nata.gov.np"},
				{Title: "Nepal's federal civil service structure explained", URL: "https://www.youtube.com/results?search_query=nepal+federal+civil+service+structure"},
			}},
		},
	}
}

func policeOfficer() careerSeed {
	return careerSeed{
		CategoryName: "Government & Law", CategorySlug: "government-law", CategoryIcon: "👮",
		Title: "Police Officer", Slug: "police-officer",
		Summary: "Police officers maintain law and order, investigate crimes, and protect citizens. They work for Nepal Police, Armed Police Force, or Provincial Police.",
		Description: "A police officer serves in law enforcement, maintaining peace and order, protecting citizens, investigating crimes, and enforcing laws. Nepal Police is the primary law enforcement agency, with the Armed Police Force handling special security situations. New provincial police forces are being developed under federalism. Police officers start at the constable level and can rise through the ranks to Inspector, DSP (Deputy Superintendent), SP (Superintendent), and higher. The Nepal Police Academy in Mahabharat, Kavre, trains recruits. Policing requires physical fitness, integrity, courage, and a commitment to justice. It is a challenging career with risks but also rewards of serving the community and maintaining law and order.",
		DailyTasks: []string{"Patrol assigned areas to deter and detect crime", "Respond to emergency calls and incidents", "Investigate crimes — collect evidence, interview witnesses, prepare reports", "Manage traffic and enforce traffic laws", "Handle public complaints and community disputes", "Prepare case files for court prosecution", "Attend court proceedings and give testimony"},
		Skills: []string{"Physical fitness and self-defense ability", "Knowledge of Nepali criminal laws (Muluki Ain, CPA)", "Investigation and evidence collection skills", "Communication and conflict de-escalation", "Report writing and documentation", "First aid and emergency response", "Integrity and ethical decision-making", "Community relations and public service orientation"},
		SalaryMin: 250000, SalaryMax: 1500000, Difficulty: 3, FutureProof: 70,
		EducationReq: "SEE pass for constable level. +2 or bachelor's for officer level (Inspector). Nepal Police Academy training. Loksewa exam for officer positions. Physical fitness test and medical examination required at all levels.",
		Outlook: "Nepal Police and Armed Police Force regularly recruit. Federalism is creating new provincial police forces. Retirements create openings. Modern policing increasingly requires technology skills (cybercrime, digital forensics). The profession offers job security and pension. Promotions depend on performance and exams.",
		Tags: []string{"government", "law enforcement", "security", "public service"},
		Resources: []resourceSeed{
			{Title: "Nepal Police Official Website", URL: "https://www.nepalpolice.gov.np", Description: "Official Nepal Police recruitment and information"},
			{Title: "Armed Police Force Nepal", URL: "https://www.apf.gov.np", Description: "Armed Police Force career opportunities"},
			{Title: "Nepal Police Academy", URL: "https://www.nepalpolice.gov.np", Description: "Police training and career development"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Meet the basic qualifications and prepare physically", Description: "Police service requires minimum educational qualifications (SEE for constable, +2/bachelor's for officer) and physical fitness. Start a fitness routine — running, push-ups, sit-ups, and endurance training. Police physical tests include running (1.6 km in 7-8 minutes), push-ups, pull-ups, and obstacle courses. Maintain a healthy lifestyle. Build general knowledge about Nepal's laws, constitution, and current affairs. Read about the police recruitment process on the Nepal Police website. Talk to serving police officers to understand the reality of the job. Policing is physically and mentally demanding.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Nepal Police Recruitment Information", URL: "https://www.nepalpolice.gov.np"},
				{Title: "Physical fitness training for police exam (YouTube)", URL: "https://www.youtube.com/results?search_query=police+physical+fitness+training+nepal"},
				{Title: "Armed Police Force recruitment process", URL: "https://www.apf.gov.np"},
			}},
			{StepNumber: 2, Title: "Apply and pass the police recruitment process", Description: "When Nepal Police or APF announces recruitment, submit your application. The selection process includes: written exam (general knowledge, Nepali, English, mathematics, and mental ability), physical fitness test, medical examination, and interview. Prepare systematically for each stage. Written exams test general knowledge, reasoning, and basic academic skills. Physical tests require you to meet minimum standards. Medical examination checks vision, hearing, and overall health. The interview assesses personality, communication, and motivation. Competition is tough — thousands apply for hundreds of positions. Prepare thoroughly and give your best.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Nepal Police written exam preparation (YouTube)", URL: "https://www.youtube.com/results?search_query=nepal+police+written+exam+preparation"},
				{Title: "Police physical test standards Nepal", URL: "https://www.nepalpolice.gov.np"},
				{Title: "Police interview questions and tips", URL: "https://www.youtube.com/results?search_query=police+interview+preparation+nepal"},
			}},
			{StepNumber: 3, Title: "Complete police academy training", Description: "Selected candidates undergo training at the Nepal Police Academy in Mahabharat, Kavre (for officers) or provincial training centers (for constables). Training lasts 6-12 months and covers: law and procedures, weapons handling, self-defense, driving, first aid, communications, ethics, and physical training. Academy life is disciplined and demanding. You will live in barracks, follow strict schedules, and undergo rigorous training. The academy transforms civilians into police officers. Build bonds with your batchmates — these relationships will support you throughout your career. Academy training is challenging but forms the foundation of your policing career.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Nepal Police Academy training life (YouTube)", URL: "https://www.youtube.com/results?search_query=nepal+police+academy+training"},
				{Title: "Police training curriculum Nepal", URL: "https://www.nepalpolice.gov.np"},
				{Title: "Weapons and self-defense training for police", URL: "https://www.youtube.com/results?search_query=police+self+defense+training+nepal"},
			}},
			{StepNumber: 4, Title: "Serve in the field and gain experience", Description: "After academy, you are posted to a police unit — district police office, traffic division, crime investigation department, or armed unit. Field experience is where you learn real policing. Every day brings new challenges — from petty disputes to serious crimes. Learn from senior officers and develop your policing style. Build relationships with the community you serve. Community policing is an important philosophy in Nepal. Maintain physical fitness and continue learning. The first few years define your career trajectory. Show initiative, integrity, and dedication. Your reputation among colleagues and superiors determines your career progress.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Daily life of a Nepal Police officer (YouTube)", URL: "https://www.youtube.com/results?search_query=life+of+nepal+police+officer"},
				{Title: "Community policing in Nepal", URL: "https://www.nepalpolice.gov.np"},
				{Title: "Crime investigation techniques (YouTube)", URL: "https://www.youtube.com/results?search_query=crime+investigation+techniques+nepal"},
			}},
			{StepNumber: 5, Title: "Prepare for promotion exams and specialize", Description: "Police career progression requires passing departmental exams for promotion. Study for promotion exams regularly. Specialize in areas like: crime investigation (CID), cybercrime, traffic management, disaster response, human trafficking investigation, or narcotics control. Specialized training is available within Nepal and through international partnerships. A specialization makes you more valuable and opens doors to specialized units. Consider higher education — a bachelor's or master's degree can help with promotions. Many officers pursue law degrees to strengthen their understanding of criminal justice. Policing is a career where continuous learning is essential.",
			Duration: "3-5 years", Links: []roadmapLink{
				{Title: "Police promotion process Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=police+promotion+process+nepal"},
				{Title: "Cybercrime investigation training Nepal", URL: "https://www.youtube.com/results?search_query=cybercrime+investigation+nepal"},
				{Title: "Nepal Police specialized training programs", URL: "https://www.nepalpolice.gov.np"},
			}},
			{StepNumber: 6, Title: "Rise to senior leadership in law enforcement", Description: "Senior police ranks in Nepal include DSP, SP, SSP, DIG, Additional IG, and Inspector General of Police (IGP). These positions involve management, policy, and leadership. Senior officers oversee districts, zones, or entire departments. Some pursue international peacekeeping missions with the UN (Nepal Police contributes to UN missions). Others move into policy roles in the Home Ministry. The IGP is the highest position in Nepal Police. Police leadership carries great responsibility — maintaining law and order is fundamental to society. Your service protects citizens and upholds the rule of law. Pride in the uniform and commitment to justice define a successful police career.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "UN peacekeeping missions for Nepal Police", URL: "https://www.youtube.com/results?search_query=nepal+police+un+peacekeeping+mission"},
				{Title: "Senior police leadership in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=nepal+police+senior+leadership"},
				{Title: "Nepal Police organizational structure", URL: "https://www.nepalpolice.gov.np"},
			}},
		},
	}
}
func judge() careerSeed {
	return careerSeed{
		CategoryName: "Government & Law", CategorySlug: "government-law", CategoryIcon: "⚖️",
		Title: "Judge", Slug: "judge",
		Summary: "Judges preside over court cases, interpret laws, and deliver justice. They ensure fair trials and uphold the rule of law in Nepal's judiciary.",
		Description: "A judge presides over court proceedings, interprets and applies laws, and delivers verdicts in civil and criminal cases. Nepal's judiciary has three tiers: District Courts, High Courts, and the Supreme Court. There are also specialized courts like the Administrative Court, Revenue Tribunal, and Labor Court. Judges are appointed by the Judicial Council and must have a law degree and legal experience. Becoming a judge is a prestigious and respected career in Nepal. Judges must be impartial, knowledgeable in law, and committed to justice. They ensure fair trials, protect the rights of citizens, and uphold constitutional values.",
		DailyTasks: []string{"Preside over court hearings and trials", "Listen to arguments from lawyers and examine evidence", "Research legal questions and precedents", "Write judgments and court orders", "Read case files and legal documents", "Manage court proceedings and courtroom staff", "Participate in judicial meetings and training"},
		Skills: []string{"Deep knowledge of Nepali laws and legal procedures", "Analytical and critical thinking", "Impartiality and ethical judgment", "Strong reading comprehension and interpretation", "Excellent writing skills for judgments", "Patience and active listening", "Courtroom management and authority", "Continuous learning of new laws and precedents"},
		SalaryMin: 600000, SalaryMax: 3600000, Difficulty: 5, FutureProof: 80,
		EducationReq: "Bachelor of Laws (LL.B) from a recognized university. Must have practiced as a lawyer for at least 7-10 years. Pass the Judicial Council exam. Master of Laws (LL.M) preferred. Nepali language proficiency essential. Candidates must be Nepali citizens of high moral character.",
		Outlook: "The judiciary offers prestigious and stable careers. District Courts have regular vacancies for judges. High Court and Supreme Court positions are highly competitive. The Judicial Council selects judges through a rigorous process. Judicial independence is constitutionally protected. Experienced judges may serve on international tribunals or become legal advisors.",
		Tags: []string{"government", "law", "justice", "prestigious"},
		Resources: []resourceSeed{
			{Title: "Judicial Council Nepal", URL: "https://www.judicialcouncil.gov.np", Description: "Judicial appointments and career information"},
			{Title: "Supreme Court of Nepal", URL: "https://www.supremecourt.gov.np", Description: "Nepal's highest court and legal resources"},
			{Title: "Nepal Bar Association", URL: "https://www.nepalbar.org.np", Description: "Legal profession and networking in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Excel in your law degree (LL.B)", Description: "The journey to becoming a judge starts with a strong legal education. Enroll in a Bachelor of Laws (LL.B) program at a recognized university in Nepal (TU, KU, Purbanchal University). Study hard — constitutional law, criminal law, civil law, evidence, jurisprudence, and procedural laws. Participate in moot court competitions to practice legal argument. Develop strong writing skills — judgments require clear, logical writing. Read Supreme Court judgments regularly to understand judicial reasoning. Build a strong foundation in legal principles. Your law school performance matters for your entire legal career.",
			Duration: "3-5 years", Links: []roadmapLink{
				{Title: "TU Law Program - Nepal Law Campus", URL: "https://www.nepallawcampus.edu.np"},
				{Title: "Supreme Court Nepal - Landmark Judgments", URL: "https://www.supremecourt.gov.np"},
				{Title: "Moot court competition tips (YouTube)", URL: "https://www.youtube.com/results?search_query=moot+court+tips+for+law+students"},
			}},
			{StepNumber: 2, Title: "Gain experience as a practicing lawyer", Description: "After LL.B, enroll in the Nepal Bar Council and become a licensed advocate. Practice law for at least 7-10 years. Work in litigation — appearing in District Courts, High Courts, and the Supreme Court. Build expertise in multiple areas of law. Develop courtroom experience, legal research skills, and client counseling. A strong practice reputation is essential for judicial appointment. Many judges spent years as successful lawyers before joining the bench. Use this time to build integrity, legal knowledge, and professional relationships. Your conduct as a lawyer will be scrutinized when you apply for judgeship.",
			Duration: "7-10 years", Links: []roadmapLink{
				{Title: "Nepal Bar Council - Advocate License", URL: "https://www.nepalbarcouncil.org.np"},
				{Title: "Building a law practice in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+build+a+law+practice+nepal"},
				{Title: "Court procedures in Nepal guide", URL: "https://www.youtube.com/results?search_query=nepal+court+procedure+guide"},
			}},
			{StepNumber: 3, Title: "Apply for judicial position through the Judicial Council", Description: "When the Judicial Council announces vacancies for judges, submit your application. The selection process includes: scrutiny of qualifications and experience, written exam (legal knowledge, judgment writing), interview, and background check. The Judicial Council assesses candidates on legal knowledge, integrity, temperament, and suitability for judicial office. Prepare thoroughly — study constitutional provisions, landmark cases, and recent legal developments. The competition is intense. Only candidates with impeccable records and deep legal knowledge are selected. The Judicial Council's selection is merit-based and transparent.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Judicial Council - Judge Selection Process", URL: "https://www.judicialcouncil.gov.np"},
				{Title: "Judicial exam preparation tips (YouTube)", URL: "https://www.youtube.com/results?search_query=judicial+exam+preparation+nepal"},
				{Title: "Role and responsibilities of judges in Nepal", URL: "https://www.youtube.com/results?search_query=role+of+judge+nepal"},
			}},
			{StepNumber: 4, Title: "Serve as a District Court judge", Description: "Newly appointed judges typically start at the District Court level. District Courts handle civil, criminal, and administrative cases at the district level. You will manage a courtroom, preside over trials, write judgments, and administer justice. District judges handle a high volume of cases. Develop efficiency without compromising quality. Learn to manage court staff and case schedules. Build a reputation for fairness, legal knowledge, and judicial temperament. District Court experience is the foundation of a judicial career. Your judgments may be reviewed by the High Court — ensure they are legally sound.",
			Duration: "3-5 years", Links: []roadmapLink{
				{Title: "District Courts of Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=district+court+nepal+working"},
				{Title: "Judgment writing techniques for judges (YouTube)", URL: "https://www.youtube.com/results?search_query=judgment+writing+for+judges"},
				{Title: "Case management in Nepali courts", URL: "https://www.youtube.com/results?search_query=case+management+nepal+court"},
			}},
			{StepNumber: 5, Title: "Advance to High Court or Supreme Court", Description: "After distinguished service as a District Judge, you may be promoted to the High Court. High Court judges hear appeals from District Courts and handle more complex cases. The highest level is the Supreme Court — Nepal's court of final appeal. Supreme Court justices interpret the constitution and set legal precedents. Advancement requires excellent judicial record, continuous legal learning, and positive recommendations from senior judges. Some judges pursue advanced legal education (LL.M, PhD) to strengthen their qualifications. The highest position is Chief Justice of Nepal. Promotion is through the Judicial Council based on merit and seniority.",
			Duration: "5-15 years", Links: []roadmapLink{
				{Title: "High Court Nepal structure and function", URL: "https://www.supremecourt.gov.np"},
				{Title: "Supreme Court Justice appointment process", URL: "https://www.judicialcouncil.gov.np"},
				{Title: "Continuing legal education for judges (YouTube)", URL: "https://www.youtube.com/results?search_query=continuing+legal+education+nepal+judges"},
			}},
			{StepNumber: 6, Title: "Serve as a senior judicial leader or legal scholar", Description: "Senior judges and justices are respected legal scholars and leaders. They may teach at law schools, write legal texts, serve on law reform commissions, or represent Nepal at international judicial conferences. Retired judges often continue contributing as arbitrators, mediators, constitutional experts, or legal advisors to the government. A judge's contribution to society extends beyond individual cases — they shape legal precedent, protect rights, and strengthen the rule of law. The judiciary is the guardian of the constitution and citizens' rights. Serving as a judge is one of the highest forms of public service.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Law reform initiatives in Nepal", URL: "https://www.supremecourt.gov.np"},
				{Title: "International judicial cooperation (YouTube)", URL: "https://www.youtube.com/results?search_query=international+judicial+cooperation+nepal"},
				{Title: "Role of judiciary in constitutional democracy Nepal", URL: "https://www.youtube.com/results?search_query=nepal+judiciary+constitutional+role"},
			}},
		},
	}
}

func diplomat() careerSeed {
	return careerSeed{
		CategoryName: "Government & Law", CategorySlug: "government-law", CategoryIcon: "🌏",
		Title: "Diplomat", Slug: "diplomat",
		Summary: "Diplomats represent Nepal abroad, manage international relations, negotiate agreements, and protect the interests of Nepali citizens overseas.",
		Description: "A diplomat serves in the Ministry of Foreign Affairs and Nepal's embassies and consulates abroad. They represent Nepal's interests internationally, negotiate treaties, promote trade and investment, provide consular services to Nepali citizens abroad, and report on political and economic developments in their host countries. Diplomats are recruited through the Loksewa Foreign Service exam. The career offers opportunities to live abroad, work on international issues, and represent Nepal on the global stage. Nepal has embassies in over 30 countries and consulates in many more. Diplomats must have excellent communication skills, international awareness, and diplomatic tact.",
		DailyTasks: []string{"Draft diplomatic correspondence, reports, and briefings", "Attend meetings with foreign officials and international organizations", "Represent Nepal at conferences, events, and ceremonies", "Process visas and provide consular services to Nepali citizens", "Analyze political and economic developments in host country", "Coordinate visits of Nepali officials abroad", "Promote Nepal's trade, tourism, and cultural interests"},
		Skills: []string{"Excellent written and verbal communication in Nepali and English", "Knowledge of international relations and diplomacy", "Negotiation and conflict resolution", "Cross-cultural understanding and adaptability", "Research and analytical skills", "Protocol and etiquette knowledge", "Foreign language ability (UN languages, Hindi, Chinese)", "Integrity, discretion, and diplomatic tact"},
		SalaryMin: 500000, SalaryMax: 3000000, Difficulty: 5, FutureProof: 65,
		EducationReq: "Master's degree preferred in International Relations, Political Science, Economics, or Law. Loksewa Foreign Service exam. Additional languages are highly valued. Training at the Foreign Service Training Academy. PhD may be required for senior positions.",
		Outlook: "Nepal's foreign service is small but active. Recruitment is through the competitive Loksewa Foreign Service exam. Opportunities expand as Nepal engages more with international organizations (UN, SAARC, BIMSTEC). Diplomats rotate between headquarters and embassies. Senior diplomats become ambassadors. The career offers prestige, international exposure, and government benefits.",
		Tags: []string{"government", "international", "foreign service", "prestigious"},
		Resources: []resourceSeed{
			{Title: "Ministry of Foreign Affairs Nepal", URL: "https://www.mofa.gov.np", Description: "Official foreign service career information"},
			{Title: "Foreign Service Training Academy Nepal", URL: "https://www.fsta.gov.np", Description: "Diplomatic training and preparation"},
			{Title: "UN Nepal", URL: "https://www.nepal.un.org", Description: "UN opportunities and Nepal's international engagement"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Build a strong academic foundation", Description: "Pursue a bachelor's and master's degree in International Relations, Political Science, Economics, or related fields. Excel academically — the Foreign Service exam is extremely competitive. Develop excellent English and Nepali language skills. Start learning additional languages — Hindi, Chinese (Mandarin), French, Arabic, or Russian. Foreign language ability significantly strengthens your candidacy. Stay informed about international affairs — read global news, follow UN developments, understand Nepal's foreign policy. Participate in Model United Nations (MUN) conferences. Build a global perspective from your student years. Diplomacy requires deep understanding of the world.",
			Duration: "4-6 years", Links: []roadmapLink{
				{Title: "MOFA Nepal - Foreign Service Information", URL: "https://www.mofa.gov.np"},
				{Title: "Model UN in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=model+un+nepal+conference"},
				{Title: "Nepal foreign policy explained", URL: "https://www.youtube.com/results?search_query=nepal+foreign+policy+explained"},
			}},
			{StepNumber: 2, Title: "Pass the Loksewa Foreign Service exam", Description: "The Foreign Service exam is conducted by the Public Service Commission. It tests: general knowledge, English, Nepali, international relations, economics, and current affairs. Prepare intensively — study international law, diplomacy theories, Nepal's foreign policy, global organizations, and current events. Join coaching institutes that offer Foreign Service preparation. The exam has multiple stages: written, interview, and language test. Only a handful of candidates are selected each year. This is one of the most difficult exams in Nepal. Prepare for years, not months. Many successful diplomats attempted the exam multiple times before being selected.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "PSC Nepal - Foreign Service Exam Information", URL: "https://www.psc.gov.np"},
				{Title: "Foreign Service exam preparation guide (YouTube)", URL: "https://www.youtube.com/results?search_query=foreign+service+exam+preparation+nepal"},
				{Title: "International relations study materials", URL: "https://www.youtube.com/results?search_query=international+relations+for+beginners"},
			}},
			{StepNumber: 3, Title: "Complete training at the Foreign Service Academy", Description: "Selected candidates undergo training at the Foreign Service Training Academy. Training covers: diplomatic protocol, consular services, international law, negotiation skills, reporting and analysis, language training, and Nepal's foreign policy framework. Learn the practical aspects of diplomacy — how embassies work, how to write diplomatic cables, how to handle official visits. Build relationships with batchmates who will be your colleagues throughout your career. Academy training transforms civilians into diplomats. Absorb everything — protocol, etiquette, communication, and the art of diplomatic representation. Your conduct as a diplomat represents Nepal.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Foreign Service Training Academy Nepal", URL: "https://www.fsta.gov.np"},
				{Title: "Diplomatic protocol and etiquette guide (YouTube)", URL: "https://www.youtube.com/results?search_query=diplomatic+protocol+etiquette+training"},
				{Title: "Consular services training for diplomats", URL: "https://www.youtube.com/results?search_query=consular+services+training+nepal"},
			}},
			{StepNumber: 4, Title: "Serve in the Ministry and at embassies abroad", Description: "Your first posting will be at the Ministry of Foreign Affairs in Kathmandu. Work in various divisions — bilateral relations, multilateral affairs, consular services, protocol, or policy planning. After initial service at headquarters, you will be posted to a Nepali embassy abroad. Diplomats typically serve 3-4 year terms abroad. Embrace the opportunity — living in another country, learning another culture, representing Nepal. Consular work (helping Nepali citizens abroad) is rewarding. Political reporting (analyzing developments in the host country) is intellectually stimulating. Each posting teaches something new.",
			Duration: "5-10 years", Links: []roadmapLink{
				{Title: "Nepali embassies and missions abroad", URL: "https://www.mofa.gov.np"},
				{Title: "Life of a Nepali diplomat abroad (YouTube)", URL: "https://www.youtube.com/results?search_query=nepali+diplomat+life+abroad"},
				{Title: "Consular services for Nepali citizens abroad", URL: "https://www.youtube.com/results?search_query=consular+services+nepal+abroad"},
			}},
			{StepNumber: 5, Title: "Develop specialization and advance in rank", Description: "Specialize in an area: bilateral relations with a specific country/region, multilateral diplomacy (UN, SAARC, BIMSTEC), international law, trade diplomacy, or consular affairs. Pursue additional training and higher education. Senior diplomats become Joint Secretaries and Ambassadors. Ambassadors are Nepal's highest diplomatic representatives in other countries. The position requires exceptional diplomatic skill, judgment, and the ability to represent Nepal's interests effectively. Ambassador positions are among the most prestigious in government service. Advancement depends on performance, experience, and continued professional development.",
			Duration: "5-10 years", Links: []roadmapLink{
				{Title: "Nepal's ambassadors and senior diplomats", URL: "https://www.mofa.gov.np"},
				{Title: "Multilateral diplomacy training (YouTube)", URL: "https://www.youtube.com/results?search_query=multilateral+diplomacy+training+online"},
				{Title: "Trade diplomacy and economic statecraft", URL: "https://www.youtube.com/results?search_query=trade+diplomacy+nepal"},
			}},
			{StepNumber: 6, Title: "Serve as a senior foreign policy leader", Description: "Top diplomats become Secretary of the Ministry of Foreign Affairs (the highest civil service position in diplomacy) or Foreign Secretary. Some serve as ambassadors to key countries (India, China, USA, UK) or permanent representatives to the United Nations. Retired diplomats contribute as foreign policy advisors, international organization officials, or academics. A diplomat's work shapes Nepal's relationships with the world. Every agreement negotiated, every crisis managed, every citizen assisted abroad — these are contributions to Nepal's place in the international community. Diplomacy builds bridges between nations.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "UN careers for Nepali diplomats", URL: "https://www.nepal.un.org"},
				{Title: "Nepal foreign secretary role and responsibilities (YouTube)", URL: "https://www.youtube.com/results?search_query=nepal+foreign+secretary+role"},
				{Title: "Career after diplomacy - international organizations", URL: "https://www.youtube.com/results?search_query=career+after+diplomacy+nepal"},
			}},
		},
	}
}
func ngoWorker() careerSeed {
	return careerSeed{
		CategoryName: "Government & Law", CategorySlug: "government-law", CategoryIcon: "🤝",
		Title: "NGO Worker (Development Professional)", Slug: "ngo-worker",
		Summary: "NGO workers design and implement development projects in health, education, livelihoods, human rights, and community development across Nepal.",
		Description: "An NGO (Non-Governmental Organization) worker or development professional works on projects that improve people's lives. Nepal has thousands of NGOs working in health, education, agriculture, livelihoods, human rights, gender equality, disaster response, and community development. International NGOs (INGOs) like Save the Children, Oxfam, UNDP, World Food Programme, and local Nepali NGOs employ development professionals. Work ranges from field-level project implementation to office-based program management, research, and advocacy. NGO careers offer the chance to make a real difference in people's lives, especially in rural and marginalized communities. The sector requires passion, dedication, and a genuine commitment to social change.",
		DailyTasks: []string{"Plan and implement development projects in communities", "Visit project sites to monitor progress and provide support", "Meet with community members, local leaders, and stakeholders", "Prepare project reports, budgets, and proposals", "Facilitate training sessions and workshops", "Coordinate with government offices and partner organizations", "Collect and analyze data to measure project impact"},
		Skills: []string{"Project planning and management (PMP, Logical Framework)", "Report writing and documentation in English and Nepali", "Community mobilization and facilitation", "Budgeting and financial management", "Monitoring and evaluation (M&E)", "Understanding of social issues in Nepal (poverty, gender, caste)", "Teamwork and cross-cultural communication", "Nepali language proficiency and local language skills"},
		SalaryMin: 300000, SalaryMax: 1800000, Difficulty: 3, FutureProof: 60,
		EducationReq: "Bachelor's in Social Work, Development Studies, Public Health, Rural Development, or related field. Master's degree preferred for senior positions. Training in project management (PMP), M&E, or gender equality. INGO experience highly valued.",
		Outlook: "Nepal has a large development sector with many INGOs, UN agencies, and local NGOs. Funding depends on donor priorities and global economic conditions. Federalism has created new opportunities in provincial and local level development work. Experienced development professionals are in demand. The sector offers good pay at senior levels, especially with INGOs and UN agencies.",
		Tags: []string{"government", "development", "social work", "community"},
		Resources: []resourceSeed{
			{Title: "Save the Children Nepal", URL: "https://www.savethechildren.org.np", Description: "Major INGO operating in Nepal"},
			{Title: "Oxfam in Nepal", URL: "https://www.oxfam.org.np", Description: "Development and humanitarian organization in Nepal"},
			{Title: "Merojob NGO Jobs", URL: "https://www.merojob.com", Description: "Find NGO/INGO jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Get education in development or social sciences", Description: "Pursue a bachelor's degree in Social Work, Development Studies, Public Health, Rural Development, Sociology, or a related field. Build understanding of Nepal's social issues — poverty, inequality, education, health, gender, caste, and regional disparities. Volunteer with local NGOs during your studies. Practical experience is as important as academic knowledge. Learn about the Sustainable Development Goals (SDGs) and how they guide development work. Read about development theories and approaches. The best NGO workers combine academic knowledge with genuine commitment to social justice and community empowerment.",
			Duration: "3-4 years", Links: []roadmapLink{
				{Title: "Development Studies programs Nepal (Edusanjal)", URL: "https://www.edusanjal.com"},
				{Title: "Sustainable Development Goals explained (YouTube)", URL: "https://www.youtube.com/results?search_query=sdg+goals+explained+nepal"},
				{Title: "Volunteering opportunities in Nepal NGOs", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 2, Title: "Volunteer and intern with NGOs", Description: "Gain practical experience by volunteering with local NGOs. Volunteer work teaches you about community dynamics, project implementation, and the realities of development work. Apply for internships with INGOs and larger Nepali NGOs. Internships provide exposure to professional development practices — proposal writing, reporting, monitoring, and evaluation. Build relationships with NGO professionals. The development sector in Nepal is relationship-driven. Your network and reputation matter for career growth. Show dedication, reliability, and genuine care for the communities you serve. Field experience in rural Nepal is especially valuable.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Volunteer in Nepal - NGO opportunities", URL: "https://www.merojob.com"},
				{Title: "Development internship programs Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=ngo+internship+nepal+guide"},
				{Title: "Community development volunteering in rural Nepal", URL: "https://www.youtube.com/results?search_query=rural+development+volunteer+nepal"},
			}},
			{StepNumber: 3, Title: "Start your professional NGO career", Description: "Apply for entry-level positions: Project Assistant, Field Officer, Monitoring Assistant, or Program Associate. Start with local NGOs or smaller INGOs. Entry-level work involves field visits, data collection, meeting coordination, and administrative support. Absorb everything — learn how projects are designed, funded, implemented, and evaluated. Understand donor requirements (USAID, DFID, UN, EU, World Bank). Develop expertise in project management tools and frameworks. Be willing to work in challenging field conditions — rural Nepal with limited amenities. Field experience is highly valued and essential for career growth.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Find NGO entry-level jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Project management basics for NGOs (YouTube)", URL: "https://www.youtube.com/results?search_query=project+management+for+ngos"},
				{Title: "Field work tips for development professionals", URL: "https://www.youtube.com/results?search_query=development+field+work+guide+nepal"},
			}},
			{StepNumber: 4, Title: "Specialize and pursue advanced education", Description: "Develop expertise in a development sector: public health, education, livelihoods, gender equality, child protection, disaster risk reduction, or climate change adaptation. Pursue a master's degree in your specialization. An MA or MSc from a good university significantly boosts your career. Get certified in key skills: Project Management (PMP), Monitoring and Evaluation (M&E), gender mainstreaming, or humanitarian standards. Specialized professionals are more valuable and earn higher salaries. Learn about donor-specific requirements and reporting formats. The development sector values continuous learning and professional development.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Public Health programs in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=public+health+programs+nepal"},
				{Title: "PMP certification for development professionals", URL: "https://www.youtube.com/results?search_query=pmp+certification+for+development+workers"},
				{Title: "Monitoring and Evaluation training (YouTube)", URL: "https://www.youtube.com/results?search_query=monitoring+and+evaluation+training+online"},
			}},
			{StepNumber: 5, Title: "Move into senior program management", Description: "With experience, move into senior roles: Program Manager, Project Coordinator, Team Leader, or Technical Advisor. Senior professionals manage teams, budgets, and multiple projects. They design programs, write proposals for funding, represent organizations to donors and government, and ensure quality and impact. At this level, strong management skills, strategic thinking, and donor relationship management are essential. INGO country office positions offer competitive salaries and benefits. Senior NGO workers in Nepal can earn 1-3 million NPR annually. Leadership in the development sector requires both technical expertise and people management skills.",
			Duration: "3-5 years", Links: []roadmapLink{
				{Title: "Program management careers in INGOs (YouTube)", URL: "https://www.youtube.com/results?search_query=program+manager+role+ingo"},
				{Title: "Proposal writing for development projects", URL: "https://www.youtube.com/results?search_query=proposal+writing+for+development+projects"},
				{Title: "Donor relationship management for NGOs", URL: "https://www.youtube.com/results?search_query=donor+management+development+sector"},
			}},
			{StepNumber: 6, Title: "Become a development leader or policy advisor", Description: "Top development professionals become Country Directors, Deputy Country Directors, or Technical Directors for INGOs and UN agencies. Some move into policy advisory roles with the Government of Nepal, bilateral donors, or international organizations. Others pursue PhDs and become researchers or academics in development studies. The most experienced professionals consult independently for multiple organizations. A career in development is deeply fulfilling — you contribute to reducing poverty, improving health, educating children, empowering women, and building a better Nepal. Every project you manage, every community you support, every life you improve — this is the reward of NGO work.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "UN careers in Nepal", URL: "https://www.nepal.un.org"},
				{Title: "Country Director roles in INGOs (YouTube)", URL: "https://www.youtube.com/results?search_query=country+director+ingo+responsibilities"},
				{Title: "Development consulting as a career", URL: "https://www.youtube.com/results?search_query=development+consulting+career+nepal"},
			}},
		},
	}
}

func socialWorker() careerSeed {
	return careerSeed{
		CategoryName: "Government & Law", CategorySlug: "government-law", CategoryIcon: "💝",
		Title: "Social Worker", Slug: "social-worker",
		Summary: "Social workers help individuals, families, and communities overcome challenges related to poverty, health, education, and social justice.",
		Description: "A social worker supports people facing challenges — poverty, abuse, disability, mental health, family problems, or social discrimination. They work in government social welfare offices, NGOs, health facilities, schools, and community organizations. Social work in Nepal addresses issues like child protection, gender-based violence, disability rights, elderly care, drug rehabilitation, and disaster psychosocial support. Social workers counsel individuals, connect people with resources, advocate for rights, and help communities build resilience. The profession requires empathy, emotional strength, and a genuine desire to help others. Social workers often work with the most vulnerable and marginalized members of society.",
		DailyTasks: []string{"Counsel individuals and families facing social challenges", "Assess client needs and develop support plans", "Connect clients with services — health, education, legal, financial", "Visit homes and communities to provide support", "Document cases and maintain confidential records", "Advocate for clients' rights and access to services", "Coordinate with other service providers and agencies"},
		Skills: []string{"Counseling and active listening skills", "Empathy and emotional resilience", "Understanding of social issues in Nepal (poverty, caste, gender)", "Case management and documentation", "Communication with diverse populations", "Problem-solving and crisis intervention", "Knowledge of social welfare policies and programs", "Self-care and boundary setting"},
		SalaryMin: 200000, SalaryMax: 900000, Difficulty: 3, FutureProof: 60,
		EducationReq: "Bachelor's in Social Work (BSW) from TU or other universities. Master's in Social Work (MSW) for senior positions. Registration with Social Welfare Council Nepal. Counseling certification and training in specialized areas (child protection, gender-based violence, mental health).",
		Outlook: "Social work is growing in Nepal as awareness of social issues increases. Government social welfare programs, NGO projects, and health facilities need social workers. Child protection, gender-based violence response, and psychosocial support are high-demand areas. The profession requires dedication — pay is modest but the impact is significant. Federalism has created new social welfare positions.",
		Tags: []string{"government", "social work", "counseling", "community"},
		Resources: []resourceSeed{
			{Title: "Social Welfare Council Nepal", URL: "https://www.swc.gov.np", Description: "Regulatory body for social welfare in Nepal"},
			{Title: "TU Social Work Program", URL: "https://www.tu.edu.np", Description: "Bachelor's and Master's in Social Work at Tribhuvan University"},
			{Title: "Merojob Social Work Jobs", URL: "https://www.merojob.com", Description: "Find social worker jobs in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Understand social issues and develop empathy", Description: "Social work starts with understanding the challenges people face. Read about social issues in Nepal — poverty, gender inequality, caste discrimination, disability, child labor, human trafficking. Volunteer at organizations that serve vulnerable populations — orphanages, disability centers, shelters for women. Listen to people's stories. Empathy is the foundation of social work. Not everyone can do this work — it requires emotional strength and genuine care for others. If you feel angry about injustice and want to help people overcome challenges, social work may be your calling. Read about Nepali social workers who have made a difference.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Social Welfare Council - Nepal", URL: "https://www.swc.gov.np"},
				{Title: "Social issues in Nepal documentary (YouTube)", URL: "https://www.youtube.com/results?search_query=social+issues+nepal+documentary"},
				{Title: "Volunteer with vulnerable communities in Nepal", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 2, Title: "Get a degree in social work (BSW)", Description: "Enroll in a Bachelor of Social Work (BSW) program at a recognized university. Study social work theories, counseling techniques, community development, social welfare policies, and research methods. BSW programs include fieldwork placements where you gain practical experience under supervision. Your fieldwork placements are opportunities to apply classroom learning and build professional skills. Build relationships with professors and fieldwork supervisors — they can guide your career. Social work education teaches you professional ethics, boundaries, and self-care. These are essential for a sustainable career helping others without burning out.",
			Duration: "3-4 years", Links: []roadmapLink{
				{Title: "TU - BSW Program Information", URL: "https://www.tu.edu.np"},
				{Title: "Social work ethics and values (YouTube)", URL: "https://www.youtube.com/results?search_query=social+work+ethics+and+values"},
				{Title: "Fieldwork in social work - tips for students", URL: "https://www.youtube.com/results?search_query=social+work+fieldwork+tips"},
			}},
			{StepNumber: 3, Title: "Start working in social welfare organizations", Description: "After BSW, apply for social worker positions in NGOs, government social welfare offices, health posts, or child protection organizations. Entry-level work includes case management, client intake, home visits, and documentation. Start building your caseload and experience. Social work is emotionally demanding. Develop self-care practices to prevent burnout. Build relationships with other service providers — health workers, police, teachers, lawyers. Effective social work requires collaboration across sectors. Each client you help builds your skills and reputation. In social work, your greatest reward is seeing lives improve.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Find social worker jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Case management in social work (YouTube)", URL: "https://www.youtube.com/results?search_query=case+management+social+work+nepal"},
				{Title: "Self-care for social workers - preventing burnout", URL: "https://www.youtube.com/results?search_query=self+care+for+social+workers"},
			}},
			{StepNumber: 4, Title: "Specialize and get advanced training", Description: "Develop expertise in a specialization: child protection, gender-based violence, mental health, medical social work, school social work, gerontology (elderly care), or disaster psychosocial support. Pursue a Master's in Social Work (MSW) to advance your career and deepen your knowledge. Get certified in specialized interventions — psychosocial counseling, trauma-informed care, family therapy, or addiction counseling. Specialized social workers are in higher demand and can earn more. Attend workshops, conferences, and training programs offered by the Social Welfare Council and NGOs. Continuous learning is essential in social work.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "MSW programs in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=msw+programs+nepal+universities"},
				{Title: "Child protection training for social workers", URL: "https://www.youtube.com/results?search_query=child+protection+training+nepal"},
				{Title: "Psychosocial counseling certification", URL: "https://www.youtube.com/results?search_query=psychosocial+counseling+training+nepal"},
			}},
			{StepNumber: 5, Title: "Move into senior or supervisory roles", Description: "Experienced social workers become senior social workers, program coordinators, or supervisors. These roles involve managing caseloads of junior workers, developing programs, training staff, and ensuring quality of services. Some social workers become trainers, teaching counseling and social work skills to others. Others move into policy and advocacy work — influencing laws and policies that affect vulnerable populations. Senior social workers in Nepal can earn 600,000-900,000 NPR annually. Management roles require both social work expertise and leadership skills. The most effective senior social workers are those who remember why they started — helping people.",
			Duration: "3-5 years", Links: []roadmapLink{
				{Title: "Social work supervision and management (YouTube)", URL: "https://www.youtube.com/results?search_query=social+work+supervision+skills"},
				{Title: "Advocacy training for social workers", URL: "https://www.youtube.com/results?search_query=advocacy+in+social+work+nepal"},
				{Title: "Social welfare policy in Nepal", URL: "https://www.swc.gov.np"},
			}},
			{StepNumber: 6, Title: "Become a leader in social welfare", Description: "Top social workers become directors of social welfare organizations, government social welfare officers, or university professors teaching social work. Some start their own NGOs addressing gaps in social services. Others become international consultants in child protection, gender equality, or social protection. Social work is more than a career — it is a commitment to social justice and human dignity. Every person you help, every family you support, every community you strengthen — these are contributions to a more just and caring Nepal. Social workers change the world one person at a time.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Social Welfare Council - Career resources", URL: "https://www.swc.gov.np"},
				{Title: "Starting an NGO in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+start+an+ngo+nepal"},
				{Title: "Social work education and academic career", URL: "https://www.youtube.com/results?search_query=social+work+academic+career+nepal"},
			}},
		},
	}
}
func carpenter() careerSeed {
	return careerSeed{
		CategoryName: "Skilled Trades", CategorySlug: "skilled-trades", CategoryIcon: "🔨",
		Title: "Carpenter", Slug: "carpenter",
		Summary: "Carpenters work with wood to build, install, and repair structures and fixtures in buildings, furniture, and other construction projects.",
		Description: "A carpenter works with wood to construct, install, and repair building structures, furniture, cabinets, doors, windows, and other wooden fixtures. Carpentry is one of the oldest and most essential skilled trades in Nepal. Carpenters work in construction (building houses, roof structures, formwork for concrete), furniture making (shops and factories), and finishing work (installing doors, windows, kitchen cabinets). Nepal's construction boom and furniture industry create steady demand for skilled carpenters. Carpentry requires precision, physical strength, and knowledge of wood types, tools, and construction techniques. Many carpenters in Nepal learned through traditional apprenticeship (training under a master carpenter). Formal vocational training is increasingly available through CTEVT.",
		DailyTasks: []string{"Measure and cut wood according to specifications", "Assemble wooden structures — frames, roofs, cabinets, furniture", "Install doors, windows, and kitchen cabinets", "Read blueprints and construction drawings", "Select appropriate wood types and materials", "Use power tools (saws, drills, sanders) safely", "Finish surfaces — sanding, staining, painting, varnishing"},
		Skills: []string{"Measurement and precision cutting", "Knowledge of wood types and their properties", "Blueprint and drawing reading", "Power tool operation and maintenance", "Joinery techniques (mortise and tenon, dovetail, etc.)", "Physical strength and stamina", "Attention to detail and quality", "Mathematics for measurements and cost estimation"},
		SalaryMin: 200000, SalaryMax: 800000, Difficulty: 3, FutureProof: 65,
		EducationReq: "SEE or +2 pass. CTEVT diploma in Furniture and Cabinet Making or General Carpentry. Apprenticeship with experienced carpenter (traditional path). Safety training and tool certification. Basic math skills essential.",
		Outlook: "Carpentry demand is steady due to construction and furniture needs. Skilled carpenters with modern tool knowledge and quality workmanship earn more and are always in demand. Specialization in kitchen cabinets, staircases, or traditional Nepali wood carving can increase earnings. Modern tools are changing the trade — carpenters who learn CNC machines and power tools have advantage.",
		Tags: []string{"skilled trades", "construction", "woodworking", "hands-on"},
		Resources: []resourceSeed{
			{Title: "CTEVT Carpentry Programs", URL: "https://www.ctevt.org.np", Description: "Vocational training in carpentry in Nepal"},
			{Title: "Merojob Carpentry Jobs", URL: "https://www.merojob.com", Description: "Find carpenter jobs in Nepal"},
			{Title: "Carpentry skills training (YouTube)", URL: "https://www.youtube.com/results?search_query=carpentry+basics+for+beginners+nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn basic carpentry skills through practice", Description: "Start by learning the basics — measuring, cutting, hammering, sawing. Practice making simple objects — a stool, a shelf, a box. Learn the names and uses of basic hand tools: hammer, saw, chisel, plane, measuring tape, square. Watch experienced carpenters at work. Many Nepali carpenters learned their trade by observing and assisting. Safety is the first lesson — learn how to use tools safely. Understand different wood types available in Nepal (sal, sisso, teak, pine, plywood) and their uses. Your first projects will be rough — that is normal. Keep practicing. Carpentry skill comes from hours of hands-on work.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Carpentry tools for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=carpentry+tools+for+beginners+nepal"},
				{Title: "CTEVT vocational training information", URL: "https://www.ctevt.org.np"},
				{Title: "Wood types in Nepal - selection guide", URL: "https://www.youtube.com/results?search_query=wood+types+in+nepal+for+furniture"},
			}},
			{StepNumber: 2, Title: "Get formal training or apprenticeship", Description: "Enroll in a CTEVT carpentry training program or find a master carpenter to apprentice under. Formal training teaches you professional techniques, safety standards, and construction mathematics. An apprenticeship with a skilled carpenter gives you real-world experience. In Nepal, many carpenters learned through the guru-shishya tradition — working under a master for years. Both paths lead to mastery. Learn to read blueprints and construction drawings. Learn modern joinery techniques. Understand how carpentry integrates with other construction trades (masonry, electrical, plumbing). A well-trained carpenter is a skilled professional who can earn good money.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "CTEVT - Furniture and Cabinet Making", URL: "https://www.ctevt.org.np"},
				{Title: "Apprenticeship benefits for carpenters (YouTube)", URL: "https://www.youtube.com/results?search_query=carpentry+apprenticeship+nepal"},
				{Title: "Blueprints and drawings for carpenters", URL: "https://www.youtube.com/results?search_query=how+to+read+blueprints+for+carpenters"},
			}},
			{StepNumber: 3, Title: "Work as a carpenter and build your skills", Description: "Start working on construction sites, furniture shops, or as a helper to an established carpenter. Accept that you will start with basic tasks — measuring, cutting, carrying materials. Show reliability and willingness to learn. Each project teaches new techniques. Build speed without compromising quality. Quality workmanship is what separates master carpenters from ordinary ones. Learn to estimate material quantities and costs — a valuable skill. Build relationships with contractors, architects, and homeowners. Repeat work and referrals come from satisfied clients. Carpentry is a skill that improves every day you work.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Find carpentry work (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Carpentry techniques - advanced joinery (YouTube)", URL: "https://www.youtube.com/results?search_query=advanced+joinery+techniques+carpentry"},
				{Title: "Construction site safety for carpenters", URL: "https://www.youtube.com/results?search_query=construction+safety+for+carpenters"},
			}},
			{StepNumber: 4, Title: "Master power tools and modern techniques", Description: "Learn to use power tools safely and efficiently — circular saw, jigsaw, planer, router, sander, drill. Power tools dramatically increase speed and precision. Learn about CNC (computer numerical control) machines used in modern furniture making. CNC operators are in high demand. Understand modern construction methods — prefabricated roof trusses, engineered wood products, and new materials. Learn about finishing techniques — stains, varnishes, paints, and sealants. A carpenter who combines traditional skills with modern tools and techniques can command premium rates. Invest in quality tools — good tools make good work possible.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Power tools for carpenters - safety guide (YouTube)", URL: "https://www.youtube.com/results?search_query=power+tools+safety+for+woodworking"},
				{Title: "CNC woodworking basics", URL: "https://www.youtube.com/results?search_query=cnc+woodworking+for+beginners"},
				{Title: "Wood finishing techniques - stain and varnish", URL: "https://www.youtube.com/results?search_query=wood+finishing+techniques+guide"},
			}},
			{StepNumber: 5, Title: "Specialize in a carpentry niche", Description: "Specialize in: furniture making (custom furniture, traditional Nepali furniture), kitchen and cabinet installation, staircases and railings, roof carpentry and formwork, restoration and antique furniture repair, or traditional Nepali wood carving (a specialized art form). Specialization allows you to charge higher prices and build a reputation. Traditional Nepali wood carving is a specialized skill with cultural significance — master carvers are highly respected. Furniture making with modern designs is in demand in urban areas. Cabinet and kitchen installation is a growing field as more Nepali homes have modern kitchens. Each specialization requires specific skills and knowledge.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "Nepali traditional wood carving (YouTube)", URL: "https://www.youtube.com/results?search_query=nepali+traditional+wood+carving+patterns"},
				{Title: "Modern kitchen cabinet making", URL: "https://www.youtube.com/results?search_query=kitchen+cabinet+making+tips+nepal"},
				{Title: "Custom furniture design and building", URL: "https://www.youtube.com/results?search_query=custom+furniture+building+techniques"},
			}},
			{StepNumber: 6, Title: "Become a master carpenter or start your own workshop", Description: "Master carpenters are those with years of experience, excellent reputation, and the ability to handle complex projects. Some start their own furniture workshop or carpentry business. Owning a workshop requires investment in tools, space, materials, and marketing. Master carpenters also teach apprentices, continuing the tradition of passing skills to the next generation. Carpentry is a respected trade that provides a good living for skilled practitioners. Nepal's construction industry and furniture market will always need good carpenters. Your hands create things that people use every day — from the roof over their heads to the furniture in their homes.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a carpentry business in Nepal", URL: "https://www.merojob.com"},
				{Title: "Master carpenter - skills and responsibilities (YouTube)", URL: "https://www.youtube.com/results?search_query=master+carpenter+skills+nepal"},
				{Title: "Training the next generation of carpenters", URL: "https://www.youtube.com/results?search_query=teaching+carpentry+to+apprentices"},
			}},
		},
	}
}

func mason() careerSeed {
	return careerSeed{
		CategoryName: "Skilled Trades", CategorySlug: "skilled-trades", CategoryIcon: "🧱",
		Title: "Mason (Brick Layer)", Slug: "mason",
		Summary: "Masons build structures with brick, block, and stone. They lay foundations, build walls, plaster surfaces, and create concrete structures for buildings and infrastructure.",
		Description: "A mason (also called brick layer or construction laborer) builds structures using bricks, concrete blocks, stones, and cement. Masonry is fundamental to construction in Nepal — most buildings use brick and cement construction. Masons lay foundations, build walls, plaster surfaces, create concrete structures (columns, beams, slabs), and install stone work. The trade requires physical strength, precision, and knowledge of materials and techniques. Masonry ranges from basic brick laying to complex stone work and decorative finishes. Experienced masons in Nepal can earn good wages, especially on larger construction projects. The trade is learned through apprenticeship — working with experienced masons on construction sites.",
		DailyTasks: []string{"Mix cement, sand, and water to prepare mortar", "Lay bricks or blocks in rows to build walls", "Apply plaster to walls and ceilings", "Build concrete foundations, columns, and beams", "Read construction drawings and follow specifications", "Use levels and plumb lines to ensure straight walls", "Install stone work and decorative finishes"},
		Skills: []string{"Knowledge of construction materials and their properties", "Brick and block laying techniques", "Plastering and finishing skills", "Concrete mixing and pouring", "Reading construction drawings", "Use of levels, plumb lines, and measuring tools", "Physical strength and stamina", "Understanding of structural basics for safety"},
		SalaryMin: 180000, SalaryMax: 700000, Difficulty: 2, FutureProof: 60,
		EducationReq: "SEE pass helpful but not required. CTEVT diploma in Masonry or Construction Technology. On-the-job training and apprenticeship is the most common path. Safety training certification. Basic math skills (measurement, calculation of materials).",
		Outlook: "Construction in Nepal is growing — housing, commercial buildings, roads, and infrastructure all need masons. Skilled masons are in steady demand. Modern construction techniques (reinforced concrete, precast) are changing the trade. Masons who learn new techniques and work efficiently earn more. Foreign employment (Gulf, Malaysia) for skilled masons is also an option.",
		Tags: []string{"skilled trades", "construction", "hands-on", "building"},
		Resources: []resourceSeed{
			{Title: "CTEVT Masonry Training", URL: "https://www.ctevt.org.np", Description: "Vocational training in masonry in Nepal"},
			{Title: "Merojob Construction Jobs", URL: "https://www.merojob.com", Description: "Find masonry and construction jobs in Nepal"},
			{Title: "Masonry techniques and tutorials (YouTube)", URL: "https://www.youtube.com/results?search_query=masonry+techniques+for+beginners+nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn basic masonry skills on construction sites", Description: "Start as a helper on a construction site. Observe experienced masons — how they mix mortar, lay bricks, check levels, and plaster walls. Learn the names and uses of masonry tools: trowel, level, plumb line, measuring tape, hammer, chisel. Practice mixing mortar (cement, sand, water ratio). Start with simple tasks — carrying materials, mixing mortar, cleaning tools. Ask questions and show eagerness to learn. Physical fitness is essential — masonry involves lifting, carrying, standing, and working in various positions. Safety awareness is critical — construction sites have many hazards.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Masonry tools and their uses (YouTube)", URL: "https://www.youtube.com/results?search_query=masonry+tools+for+beginners+nepal"},
				{Title: "Construction site safety basics (YouTube)", URL: "https://www.youtube.com/results?search_query=construction+site+safety+basics+nepal"},
				{Title: "CTEVT construction training programs", URL: "https://www.ctevt.org.np"},
			}},
			{StepNumber: 2, Title: "Get formal training or apprenticeship", Description: "Enroll in a CTEVT masonry or construction technology course. Formal training teaches proper techniques, material science, and safety standards. Or continue the apprenticeship path — work under a master mason who teaches you the trade. Learn wall types (load-bearing, partition, retaining), bond patterns (English, Flemish, stretcher), and concrete work (mix ratios, pouring, curing). Learn to read basic construction drawings. Understand foundation types and their construction. Knowledge of materials — bricks, blocks, stones, cement, sand, aggregate, reinforcement steel. A trained mason with theoretical knowledge is more valuable than one who only knows practical techniques.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "CTEVT - Diploma in Construction Technology", URL: "https://www.ctevt.org.np"},
				{Title: "Brick bond patterns explained (YouTube)", URL: "https://www.youtube.com/results?search_query=brick+bond+patterns+explained+english+bond"},
				{Title: "Concrete mix ratios and curing methods", URL: "https://www.youtube.com/results?search_query=concrete+mix+ratio+guide+nepal"},
			}},
			{StepNumber: 3, Title: "Work as a skilled mason on construction projects", Description: "Start working as a mason on housing projects, commercial buildings, or infrastructure work. Each project type teaches different skills. Residential work involves foundations, walls, plaster, and finishing. Commercial work includes larger structures, reinforced concrete, and complex designs. Develop speed without sacrificing quality. A good mason produces straight walls, even plaster, and strong structures. Learn to work efficiently — organize materials, minimize waste, and complete work on schedule. Build a reputation for reliability and quality work. Word of mouth is how masons get jobs in Nepal. Satisfied clients and contractors will call you for future projects.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Find masonry work (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Advanced brick laying techniques (YouTube)", URL: "https://www.youtube.com/results?search_query=advanced+brick+laying+techniques"},
				{Title: "Plastering tips for smooth walls", URL: "https://www.youtube.com/results?search_query=plastering+tips+for+smooth+finish+nepal"},
			}},
			{StepNumber: 4, Title: "Learn modern construction methods", Description: "Construction technology is evolving. Learn reinforced cement concrete (RCC) work — column and beam construction, slab pouring, and formwork. Learn about precast concrete elements used in modern construction. Understand earthquake-resistant construction techniques — very important in Nepal's seismic zone. Learn waterproofing methods for roofs and basements. Knowledge of modern materials (hollow blocks, aerated concrete, stone cladding) gives you an edge. Learn about construction chemicals — admixtures, sealants, bonding agents. A mason who understands modern techniques and earthquake safety is more valuable and can charge more for their skills.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Earthquake-resistant building construction (YouTube)", URL: "https://www.youtube.com/results?search_query=earthquake+resistant+building+nepal"},
				{Title: "RCC column and beam construction methods", URL: "https://www.youtube.com/results?search_query=rcc+column+beam+construction+nepal"},
				{Title: "Waterproofing techniques for buildings", URL: "https://www.youtube.com/results?search_query=waterproofing+techniques+for+roof+nepal"},
			}},
			{StepNumber: 5, Title: "Specialize in decorative or high-end finishes", Description: "Specialize in: decorative brickwork, stone masonry (natural stone walls, paving), tile and marble installation, restoration masonry (heritage buildings), or plaster finishes (textured, smooth, decorative). Stone masonry is valuable for retaining walls, landscaping, and decorative features. Tile and marble work for bathrooms and kitchens is in high demand in urban areas. Heritage restoration masonry is specialized work for Nepal's historic temples and palaces. Specialized masons earn significantly more than general masons. Each specialization requires additional training and practice.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "Stone masonry techniques (YouTube)", URL: "https://www.youtube.com/results?search_query=stone+masonry+techniques+nepal"},
				{Title: "Tile installation and marble work guide", URL: "https://www.youtube.com/results?search_query=tile+installation+guide+nepal"},
				{Title: "Heritage restoration masonry Nepal", URL: "https://www.youtube.com/results?search_query=heritage+building+restoration+nepal"},
			}},
			{StepNumber: 6, Title: "Become a master mason or construction supervisor", Description: "Master masons are highly respected in Nepal's construction industry. They supervise teams of masons and laborers, ensure quality, and handle complex projects. Some start their own construction contracting businesses. Construction supervisors coordinate all aspects of building projects, managing multiple trades. Masonry skills also provide opportunities for foreign employment in Gulf countries and Malaysia, where skilled masons earn higher wages. A skilled mason builds the structures that shelter families, businesses, and communities. Your work is literally the foundation of Nepal's built environment. Take pride in quality workmanship — every wall you build should stand for generations.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Construction supervision and management (YouTube)", URL: "https://www.youtube.com/results?search_query=construction+supervision+nepal"},
				{Title: "Starting a contracting business in Nepal", URL: "https://www.merojob.com"},
				{Title: "Foreign employment for skilled Nepali masons", URL: "https://www.youtube.com/results?search_query=nepali+masons+foreign+employment+gulf"},
			}},
		},
	}
}
func tailor() careerSeed {
	return careerSeed{
		CategoryName: "Skilled Trades", CategorySlug: "skilled-trades", CategoryIcon: "✂️",
		Title: "Tailor", Slug: "tailor",
		Summary: "Tailors design, sew, and alter clothing. They create custom garments, repair clothes, and provide fitting services for individual clients.",
		Description: "A tailor makes, alters, and repairs clothing for individual customers. Tailoring is an essential trade in Nepal — from traditional dress (kurta suruwal, daura suruwal, sari blouses, lehenga) to Western-style suits, shirts, pants, and jackets. Tailors work in small shops, larger tailoring businesses, or from home. The garment industry in Nepal (ready-made garments, pashmina, handicraft clothing) also employs tailors. Traditionally, tailoring skills were passed down through families. Today, CTEVT offers formal tailoring programs. Tailoring can be a good self-employment career with flexibility. Skilled tailors who produce quality garments earn good money and have loyal customers. The trade requires patience, precision, and an eye for detail.",
		DailyTasks: []string{"Take customer measurements for custom garments", "Cut fabric according to patterns and measurements", "Sew garments using sewing machines or hand stitching", "Alter existing clothes — hemming, taking in, letting out", "Fit garments on customers and make adjustments", "Repair torn or damaged clothing", "Manage tailoring shop — billing, orders, customer relations"},
		Skills: []string{"Expert sewing machine operation", "Pattern making and drafting", "Measurement and fitting skills", "Knowledge of fabrics and their properties", "Alteration and repair techniques", "Customer service and communication", "Business management for self-employed tailors", "Design sense and understanding of fashion trends"},
		SalaryMin: 150000, SalaryMax: 600000, Difficulty: 2, FutureProof: 50,
		EducationReq: "SEE pass. CTEVT diploma in Tailoring and Garment Technology. Apprenticeship with experienced tailor. Basic math for measurements and billing. Design and fashion courses from institutes like Janakpur Handicraft or CTEVT.",
		Outlook: "Tailoring demand remains steady — people always need clothes altered and custom garments made. Ready-made garments compete but custom tailoring is preferred for special occasions (weddings, festivals) and traditional dress. The garment industry in Nepal provides factory-based tailoring jobs. Tailoring provides good self-employment opportunities. Fashion awareness is increasing, creating demand for skilled tailors.",
		Tags: []string{"skilled trades", "garment", "fashion", "hands-on"},
		Resources: []resourceSeed{
			{Title: "CTEVT Tailoring Programs", URL: "https://www.ctevt.org.np", Description: "Vocational training in tailoring and garment technology"},
			{Title: "Merojob Garment Jobs", URL: "https://www.merojob.com", Description: "Find tailoring and garment jobs in Nepal"},
			{Title: "Sewing techniques for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=sewing+techniques+for+beginners+nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn basic sewing skills", Description: "Start with hand sewing — learn stitches (running stitch, backstitch, hem stitch). Learn to use a sewing machine — practice straight lines, curves, and different stitch types. Practice on simple projects: pillowcases, bags, simple skirts or pants. Learn fabric types — cotton, silk, polyester, denim, traditional Nepali fabrics (dhaka, hemp, khadi). Understanding fabric behavior is essential for quality tailoring. Take measurements correctly — measurement errors lead to ill-fitting garments. Practice on family members first. Watch experienced tailors at work. Sewing is a skill that improves with hours of practice.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Sewing machine basics for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=sewing+machine+basics+for+beginners"},
				{Title: "Fabric types and their uses explained", URL: "https://www.youtube.com/results?search_query=fabric+types+guide+for+tailors"},
				{Title: "How to take body measurements correctly", URL: "https://www.youtube.com/results?search_query=how+to+take+body+measurements+for+tailoring"},
			}},
			{StepNumber: 2, Title: "Get formal training or apprenticeship", Description: "Enroll in a CTEVT tailoring program or apprentice with an experienced tailor. Formal training teaches pattern making, drafting, grading (sizing patterns), and professional finishing techniques. You will learn garment construction for different types of clothing — shirts, pants, dresses, suits, traditional wear. Apprenticeship gives you real customer experience — handling fittings, dealing with difficult fabrics, and managing a tailoring business. Learn about traditional Nepali garments — daura suruwal, kurta suruwal, sari blouse, lehenga choli, gunyu cholo. These garments have specific construction requirements that every Nepali tailor should know. Specialized knowledge of traditional wear is valuable in the Nepal market.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "CTEVT - Garment Technology Program", URL: "https://www.ctevt.org.np"},
				{Title: "Pattern making for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=pattern+making+for+beginners+tailoring"},
				{Title: "Traditional Nepali garments - making guide", URL: "https://www.youtube.com/results?search_query=traditional+nepali+dress+making"},
			}},
			{StepNumber: 3, Title: "Start working as a tailor", Description: "Work in a tailoring shop, garment factory, or start taking orders from home. Build speed while maintaining quality. Each garment you make teaches you something. Learn to handle different customer requests — some want exact copies of ready-made garments, others want custom designs. Develop communication skills to understand what customers want. Sometimes customers cannot articulate what they want — a good tailor guides them. Build a portfolio of your work — photographs of garments you have made. Word of mouth is how tailors build their business. Satisfied customers bring family, friends, and colleagues. Quality, reliability, and fair pricing build a loyal customer base.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Find tailoring jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Customer service tips for tailors (YouTube)", URL: "https://www.youtube.com/results?search_query=customer+service+for+tailors"},
				{Title: "Building a tailoring portfolio", URL: "https://www.youtube.com/results?search_query=build+tailoring+portfolio+guide"},
			}},
			{StepNumber: 4, Title: "Learn advanced techniques and specialty areas", Description: "Master advanced tailoring: suit making (jackets, blazers), bridal wear (wedding dresses, elaborate traditional wear), leather garments, and intricate embroidery. Learn about fashion trends and update your skills accordingly. Learn about different fashion styles and how to advise customers on what suits them. Understand fit — how garments should fit different body types. Good fit is the difference between an ordinary tailor and an excellent one. Learn about alterations — many tailors earn significant income from alteration work alone. Alterations require different skills than making garments from scratch. Master both areas.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Suit making and tailoring techniques (YouTube)", URL: "https://www.youtube.com/results?search_query=suit+making+techniques+for+tailors"},
				{Title: "Bridal wear design and construction", URL: "https://www.youtube.com/results?search_query=bridal+wear+making+guide+nepal"},
				{Title: "Garment fitting and alteration techniques", URL: "https://www.youtube.com/results?search_query=garment+fitting+and+alteration+tips"},
			}},
			{StepNumber: 5, Title: "Open your own tailoring shop or specialize", Description: "Opening a tailoring shop requires investment in sewing machines (industrial machines for faster production), space, and initial fabric stock. Choose a location with good visibility and foot traffic. Offer additional services: dry cleaning, alterations, fabric sales. These complementary services increase customer traffic. Specialize in a niche: bridal wear, men's suits, traditional Nepali wear, or leather garments. Specialization allows you to build a reputation as an expert and charge premium prices. Build relationships with fabric shops, bridal shops, and other businesses that can refer customers. A well-run tailoring shop provides a stable income.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Starting a tailoring business in Nepal", URL: "https://www.merojob.com"},
				{Title: "Industrial sewing machines for tailoring shops (YouTube)", URL: "https://www.youtube.com/results?search_query=industrial+sewing+machine+for+tailoring+business"},
				{Title: "Tailoring business marketing tips", URL: "https://www.youtube.com/results?search_query=tailoring+shop+marketing+nepal"},
			}},
			{StepNumber: 6, Title: "Expand into garment manufacturing or fashion design", Description: "Experienced tailors can expand into garment manufacturing — producing ready-made garments for the Nepal market or export. The garment industry in Nepal includes pashmina products, ready-made garments, and traditional handicraft clothing. Some tailors study fashion design and become fashion designers — creating original collections and participating in fashion shows. Others teach tailoring at vocational training institutes, passing their skills to the next generation. Tailoring is a skill that provides lifelong career security. People will always need clothes. A skilled tailor who produces quality work will never be without customers.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Garment manufacturing business in Nepal", URL: "https://www.youtube.com/results?search_query=garment+manufacturing+nepal+guide"},
				{Title: "Fashion design education in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=fashion+design+education+nepal"},
				{Title: "Teaching tailoring - vocational training career", URL: "https://www.ctevt.org.np"},
			}},
		},
	}
}

func evMechanic() careerSeed {
	return careerSeed{
		CategoryName: "Skilled Trades", CategorySlug: "skilled-trades", CategoryIcon: "⚡",
		Title: "EV Mechanic (Electric Vehicle Technician)", Slug: "ev-mechanic",
		Summary: "EV mechanics diagnose, repair, and maintain electric vehicles — including e-rickshaws, electric scooters, e-bikes, and electric cars.",
		Description: "An electric vehicle (EV) mechanic specializes in servicing and repairing electric vehicles. As Nepal rapidly adopts electric mobility — e-rickshaws, electric scooters (Neta, Yatri, etc.), and increasingly electric cars — the demand for skilled EV mechanics is exploding. EV mechanics work with high-voltage batteries, electric motors, controllers, and charging systems. Unlike traditional mechanics who work on engines, EV mechanics work with electrical and electronic systems. This is a future-proof career with growing demand as Nepal targets 90% EV penetration by 2030. EV mechanics need electrical knowledge more than mechanical skills. The government's EV promotion policies, tax incentives, and import trends all point to massive growth in this field.",
		DailyTasks: []string{"Diagnose electrical faults using diagnostic tools and multimeters", "Service and replace high-voltage batteries", "Repair or replace electric motors and controllers", "Install and maintain EV charging equipment", "Perform regular maintenance — brake systems, tires, suspension", "Update vehicle software and firmware", "Advise customers on EV care and battery management"},
		Skills: []string{"Electrical and electronics knowledge (DC/AC circuits, voltage, current)", "Diagnostic tool operation (multimeter, oscilloscope, scan tools)", "High-voltage battery service and safety", "Electric motor and controller repair", "EV charging system knowledge", "Mechanical skills for general vehicle maintenance", "Customer education on EV care", "Safety procedures for high-voltage systems"},
		SalaryMin: 240000, SalaryMax: 1200000, Difficulty: 3, FutureProof: 90,
		EducationReq: "CTEVT diploma in Electrical and Electronics or EV Technology. Training from EV manufacturers (Yatri, Tesla, etc.). Certification in high-voltage safety. Mechanical background helpful but electrical knowledge is essential. Many traditional mechanics are reskilling into EV repair.",
		Outlook: "EV adoption in Nepal is accelerating rapidly. The government aims for 90% EVs by 2030. E-rickshaws are ubiquitous in the Terai. Electric scooters are dominating Kathmandu sales. EV mechanics are in short supply — this is a high-demand, high-growth career. Early entrants in this field can build successful businesses. As EVs grow, demand for skilled technicians will only increase.",
		Tags: []string{"skilled trades", "electric vehicle", "automotive", "future-proof"},
		Resources: []resourceSeed{
			{Title: "CTEVT EV Technology Programs", URL: "https://www.ctevt.org.np", Description: "Vocational training in electric vehicle technology"},
			{Title: "Yatri Motorcycles Nepal", URL: "https://www.yatri.com.np", Description: "Nepali EV manufacturer and service network"},
			{Title: "EV Nepal news and resources", URL: "https://www.youtube.com/results?search_query=ev+repair+nepal+guide"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn the basics of electricity and electronics", Description: "EV mechanics need strong electrical knowledge. Start with basics — voltage, current, resistance, power, DC and AC circuits. Learn about batteries — how lithium-ion batteries work, battery management systems (BMS), charging cycles, and safety. Learn to use a multimeter (essential for diagnostics). Understand electrical diagrams and circuit tracing. Study basic electronics — controllers, sensors, motors, inverters. If you have a background in traditional vehicle mechanics, recognize that EV repair requires different skills. You are working with high voltage (up to 400V) that can be lethal — safety training is critical.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Basic electronics for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=basic+electronics+for+beginners+nepal"},
				{Title: "Multimeter guide for beginners", URL: "https://www.youtube.com/results?search_query=how+to+use+multimeter+for+beginners"},
				{Title: "Lithium-ion battery basics explained", URL: "https://www.youtube.com/results?search_query=lithium+ion+battery+working+principle"},
			}},
			{StepNumber: 2, Title: "Get formal EV training", Description: "Enroll in an EV technology program at CTEVT or a technical training institute. Learn about EV systems: powertrain, battery pack, battery management system, motor controller, DC-DC converter, onboard charger, and thermal management. Learn high-voltage safety procedures — insulated tools, PPE, emergency shutdown procedures. Understand the differences between various EV types — e-rickshaws, e-scooters, e-bikes, e-cars, e-rickshaws. Each type has different systems and service requirements. Manufacturer-specific training is also valuable — Yatri, Neta, MG, and other brands offer technician training. Practical hands-on training is essential.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "CTEVT - EV Technician Program", URL: "https://www.ctevt.org.np"},
				{Title: "High voltage safety training for EV technicians (YouTube)", URL: "https://www.youtube.com/results?search_query=high+voltage+safety+ev+technician"},
				{Title: "EV powertrain components explained", URL: "https://www.youtube.com/results?search_query=ev+powertrain+components+explained"},
			}},
			{StepNumber: 3, Title: "Work as an EV technician apprentice", Description: "Join an EV dealership service center, EV repair shop, or a garage that services EVs. Start with basic tasks — battery pack removal and installation, controller diagnostics, motor testing. Learn from experienced EV technicians. EV technology is evolving rapidly — every new model brings new systems and components. Stay updated through manufacturer training and online resources. Build diagnostic skills — EV faults are often electrical rather than mechanical, requiring systematic troubleshooting. Learn to use diagnostic software — modern EVs have complex electronic control units (ECUs) that require software diagnostics. Good diagnostic skills make you a valuable technician.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Find EV mechanic jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "EV diagnostic tools and techniques (YouTube)", URL: "https://www.youtube.com/results?search_query=ev+diagnostic+tools+guide"},
				{Title: "Common EV faults and troubleshooting", URL: "https://www.youtube.com/results?search_query=common+ev+faults+and+solutions"},
			}},
			{StepNumber: 4, Title: "Specialize in a specific EV type or brand", Description: "Specialize in: e-rickshaw repair (massive market in Terai cities), electric scooter/motorcycle service (growing rapidly in Kathmandu), electric car service (long-term growth), or battery pack rebuilding and repair. Battery repair and rebuilding is a specialized high-value skill. EV batteries are expensive and often repairable rather than replaceable. Learn about battery cell replacement, BMS repair, and battery balancing. Specialization in battery repair can be very profitable. EV charger installation is another growing specialty — installing home and public charging stations. Each specialization requires additional training but commands higher pay.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "E-rickshaw repair and maintenance (YouTube)", URL: "https://www.youtube.com/results?search_query=e+rickshaw+repair+nepal"},
				{Title: "EV battery pack repair and rebuilding", URL: "https://www.youtube.com/results?search_query=ev+battery+pack+repair+guide"},
				{Title: "EV charger installation training", URL: "https://www.youtube.com/results?search_query=ev+charger+installation+guide"},
			}},
			{StepNumber: 5, Title: "Open your own EV service center", Description: "As EV adoption grows, dedicated EV service centers are needed. Start a specialized EV repair shop. Investment needed: diagnostic equipment, high-voltage safety gear, battery handling equipment, and specialized tools. Location near areas with high EV usage (Kathmandu valley, Terai highway towns). Build relationships with EV dealers and insurance companies for referral business. Train your staff — good EV mechanics are hard to find. Offer mobile EV repair service — go to customers who cannot bring their EVs to you. Mobile EV repair is an underserved market. EV service centers have strong growth potential as EV numbers increase exponentially.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "Starting an EV repair business in Nepal", URL: "https://www.youtube.com/results?search_query=ev+repair+business+nepal"},
				{Title: "EV service center equipment guide", URL: "https://www.youtube.com/results?search_query=ev+service+center+equipment+needed"},
				{Title: "Mobile EV repair business model", URL: "https://www.youtube.com/results?search_query=mobile+ev+repair+service+business"},
			}},
			{StepNumber: 6, Title: "Become an EV technology expert and trainer", Description: "Top EV technicians become trainers, teaching EV repair at technical institutes. Some join EV manufacturers as technical specialists or quality assurance managers. Others become EV conversion specialists — converting conventional vehicles to electric. EV conversion is an emerging field with significant potential. As EV technology evolves (solid-state batteries, wireless charging, V2G), continuous learning is essential. EV technicians who stay at the forefront of technology will always be in demand. The EV revolution in Nepal is creating opportunities for a whole generation of skilled technicians. Being an early adopter of EV skills positions you perfectly for the future of transportation.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "EV conversion - converting cars to electric (YouTube)", URL: "https://www.youtube.com/results?search_query=ev+conversion+kit+nepal"},
				{Title: "Teaching EV technology in Nepal", URL: "https://www.ctevt.org.np"},
				{Title: "Future of EV technology - what to learn next", URL: "https://www.youtube.com/results?search_query=future+of+ev+technology+trends"},
			}},
		},
	}
}
func motorcycleMechanic() careerSeed {
	return careerSeed{
		CategoryName: "Skilled Trades", CategorySlug: "skilled-trades", CategoryIcon: "🏍️",
		Title: "Motorcycle Mechanic", Slug: "motorcycle-mechanic",
		Summary: "Motorcycle mechanics diagnose, repair, and maintain motorcycles, scooters, and other two-wheelers. They work on engines, electrical systems, brakes, and more.",
		Description: "A motorcycle mechanic services and repairs motorcycles, scooters, and mopeds. With millions of two-wheelers on Nepal's roads, motorcycle mechanics are always in demand. They work at dealership service centers, independent garages, or run their own repair shops. Motorcycle mechanics diagnose engine problems, repair brakes and suspension, fix electrical issues, service transmissions, and perform routine maintenance. The job requires mechanical aptitude, problem-solving skills, and physical stamina. Nepal has many motorcycle brands (Hero, Honda, Bajaj, Yamaha, Suzuki, TVS) and each requires specific knowledge. Two-wheelers are the most common personal vehicles in Nepal, especially in rural and hilly areas where roads are narrow. Mechanics who build a reputation for honest, quality work have steady customers.",
		DailyTasks: []string{"Diagnose mechanical and electrical problems in motorcycles", "Repair or replace engines, transmissions, and clutches", "Service brake systems, suspension, and steering", "Repair electrical systems — wiring, batteries, lights, starters", "Perform routine maintenance — oil change, chain adjustment, tune-ups", "Test ride motorcycles after repairs to ensure quality", "Order parts and manage inventory for the workshop"},
		Skills: []string{"Engine repair and overhaul (2-stroke and 4-stroke)", "Electrical system diagnostics and repair", "Brake, suspension, and steering system knowledge", "Use of diagnostic tools and workshop equipment", "Knowledge of different motorcycle brands and models", "Customer service and communication", "Parts identification and inventory management", "Physical fitness and mechanical aptitude"},
		SalaryMin: 180000, SalaryMax: 700000, Difficulty: 2, FutureProof: 50,
		EducationReq: "SEE pass. CTEVT diploma in Motorcycle Mechanics or Automotive Engineering. Apprenticeship with experienced mechanic. Brand-specific training from manufacturers (Hero, Honda, Bajaj). Safety training. Basic math and measurements.",
		Outlook: "Motorcycles will remain the dominant vehicle in Nepal for the foreseeable future. Demand for mechanics is steady. Brand-specific knowledge (diagnostic computers, specialized tools) is increasingly important as motorcycles become more technologically advanced. Electric motorcycles are emerging (Yatri, etc.) — mechanics who learn both conventional and EV two-wheeler repair will have the best prospects.",
		Tags: []string{"skilled trades", "automotive", "motorcycle", "hands-on"},
		Resources: []resourceSeed{
			{Title: "CTEVT Automotive Programs", URL: "https://www.ctevt.org.np", Description: "Vocational training in automotive and motorcycle mechanics"},
			{Title: "Merojob Mechanic Jobs", URL: "https://www.merojob.com", Description: "Find motorcycle mechanic jobs in Nepal"},
			{Title: "Motorcycle repair tutorials (YouTube)", URL: "https://www.youtube.com/results?search_query=motorcycle+repair+for+beginners+nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn basic mechanical skills on motorcycles", Description: "Start by working on your own motorcycle or helping friends with theirs. Learn basic maintenance: oil change, chain adjustment, brake adjustment, tire change, air filter cleaning. Understand how a motorcycle works — engine (2-stroke vs 4-stroke), transmission, clutch, brakes, suspension, electrical system. Learn the names and uses of tools: wrenches, screwdrivers, pliers, sockets, torque wrench, feeler gauge. Spend time in a workshop observing experienced mechanics. Ask questions. Many of Nepal's best motorcycle mechanics learned by apprenticeship — working alongside a senior mechanic for years. Start with small tasks and work your way up to complex repairs.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Basic motorcycle maintenance guide (YouTube)", URL: "https://www.youtube.com/results?search_query=basic+motorcycle+maintenance+guide+nepal"},
				{Title: "Motorcycle tools and workshop setup", URL: "https://www.youtube.com/results?search_query=motorcycle+repair+tools+for+beginners"},
				{Title: "How a motorcycle engine works - explained", URL: "https://www.youtube.com/results?search_query=how+motorcycle+engine+works+animation"},
			}},
			{StepNumber: 2, Title: "Get formal training or brand certification", Description: "Enroll in a CTEVT motorcycle mechanics or automotive engineering program. Formal training teaches systematic diagnostics, engine overhaul, electrical systems, and professional workshop practices. Get brand-specific training from motorcycle manufacturers (Hero, Honda, Bajaj, Yamaha). Brand training centers in Nepal offer certified courses. Certification from a manufacturer increases your employability at authorized service centers. Learn about fuel systems (carburetor and fuel injection), ignition systems, and modern engine management. Modern motorcycles have complex electrical systems and in some cases, ECU diagnostics. Training keeps you current with evolving technology.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "CTEVT - Automotive Mechanics Program", URL: "https://www.ctevt.org.np"},
				{Title: "Honda motorcycle service training (YouTube)", URL: "https://www.youtube.com/results?search_query=honda+motorcycle+service+training+nepal"},
				{Title: "Motorcycle engine overhaul - step by step", URL: "https://www.youtube.com/results?search_query=motorcycle+engine+overhaul+guide"},
			}},
			{StepNumber: 3, Title: "Work as a mechanic at a garage or service center", Description: "Start working at a motorcycle dealership service center, a roadside garage, or a specialized repair shop. Dealership service centers offer stable employment, training, and access to genuine parts. Independent garages offer variety — you will work on many brands and types of repairs. Build diagnostic skills — diagnosing faults quickly and accurately is the most valuable mechanic skill. Develop speed without compromising quality. Build a reputation for honesty — customers trust mechanics who are honest about what needs repair and what does not. Word of mouth is everything in this business. Honest, skilled mechanics have lifelong customers.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Find mechanic jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Motorcycle diagnostic techniques (YouTube)", URL: "https://www.youtube.com/results?search_query=motorcycle+diagnostic+techniques+guide"},
				{Title: "Customer trust and honesty in auto repair", URL: "https://www.youtube.com/results?search_query=honest+auto+repair+business+tips"},
			}},
			{StepNumber: 4, Title: "Master advanced repair and specialty areas", Description: "Learn advanced areas: engine rebuilding (re-boring, valve grinding, crank shaft repair), electrical system design and repair (rewiring, ECU diagnostics, ABS systems), fuel injection systems (modern motorcycles are moving from carburetor to EFI), and performance modification (engine tuning, suspension setup). Master diagnostic equipment — compression tester, vacuum gauge, timing light, multimeter, OBD scanner. Specialize in: vintage motorcycle restoration (many classic motorcycles in Nepal), performance tuning, or a specific brand. Advanced skills command higher rates. A mechanic who can fix any motorcycle — from a Hero Splendor to a Harley Davidson — is always in demand.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "Engine rebuilding and re-boring (YouTube)", URL: "https://www.youtube.com/results?search_query=motorcycle+engine+rebuilding+nepal"},
				{Title: "Motorcycle electrical system repair guide", URL: "https://www.youtube.com/results?search_query=motorcycle+electrical+system+troubleshooting"},
				{Title: "Fuel injection vs carburetor - service guide", URL: "https://www.youtube.com/results?search_query=fuel+injection+system+repair+motorcycle"},
			}},
			{StepNumber: 5, Title: "Open your own motorcycle repair shop", Description: "Starting your own workshop requires: space (preferably on a busy road with high two-wheeler traffic), tools and equipment (lift, compressor, diagnostic tools, specialty tools for different brands), parts inventory (oil, filters, brake pads, cables, common spare parts), business registration from local municipality, and a signboard. Build relationships with spare parts dealers. Good parts sourcing at competitive prices gives you an edge. Offer services beyond repair: custom painting, accessory installation, motorcycle detailing, and performance upgrades. A well-located, well-equipped shop with a good reputation can generate steady income.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Starting a motorcycle garage in Nepal", URL: "https://www.youtube.com/results?search_query=start+motorcycle+garage+nepal"},
				{Title: "Motorcycle repair shop equipment guide", URL: "https://www.youtube.com/results?search_query=workshop+equipment+for+motorcycle+repair"},
				{Title: "Spare parts sourcing for motorcycle shops Nepal", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 6, Title: "Expand business or transition to training/management", Description: "Expanding your business: add more service bays, hire and train mechanics, open multiple locations, or become a spare parts distributor. Experienced mechanics become service managers at dealerships, workshop supervisors, or technical trainers at CTEVT institutes. Some become specialized in automotive electrical, becoming the go-to expert for wiring and ECU problems. Motorcycles are central to transportation in Nepal. A skilled and honest mechanic has job security and the satisfaction of keeping Nepal moving. Your skills help people commute to work, transport goods, and travel across the country.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Growing a garage business - multiple locations", URL: "https://www.youtube.com/results?search_query=growing+auto+repair+business"},
				{Title: "Technical training career for mechanics", URL: "https://www.ctevt.org.np"},
				{Title: "Motorcycle service management tips", URL: "https://www.youtube.com/results?search_query=motorcycle+service+center+management"},
			}},
		},
	}
}

func welder() careerSeed {
	return careerSeed{
		CategoryName: "Skilled Trades", CategorySlug: "skilled-trades", CategoryIcon: "🔥",
		Title: "Welder", Slug: "welder",
		Summary: "Welders join metal parts together using heat and pressure. They work on construction, manufacturing, fabrication, and repair projects across industries.",
		Description: "A welder joins metal components using heat (from electric arc, gas flame, or other sources) to melt and fuse the pieces together. Welding is an essential skill in construction (steel structures, rebar, railings), manufacturing (factory equipment, metal products), fabrication (gates, windows, furniture), automotive repair, and infrastructure (bridges, pipelines). Nepal has growing demand for welders in construction, hydropower projects, manufacturing, and repair work. Skilled welders who produce strong, clean welds are always in demand. Welding requires steady hands, good eyesight (with proper eye protection), knowledge of metal types, and safety consciousness. Welders can work as employees or start their own fabrication business.",
		DailyTasks: []string{"Set up welding equipment and select appropriate settings", "Weld metal components together following specifications", "Cut metal using grinders, torches, or plasma cutters", "Grind and finish welds for smooth appearance", "Read blueprints and fabrication drawings", "Inspect welds for quality and strength", "Maintain welding equipment and safety gear"},
		Skills: []string{"Arc welding (SMAW/stick), MIG, TIG welding techniques", "Metal cutting and grinding", "Blueprint and drawing reading", "Knowledge of different metals (steel, aluminum, stainless steel)", "Measurement and precision", "Safety procedures and PPE use", "Physical strength and stamina", "Attention to detail for quality welds"},
		SalaryMin: 200000, SalaryMax: 800000, Difficulty: 3, FutureProof: 65,
		EducationReq: "SEE pass. CTEVT diploma in Welding and Fabrication. Certification in specific welding techniques (Arc, MIG, TIG). Safety training (OSHA or equivalent). Apprenticeship with experienced welder. Basic math for measurements.",
		Outlook: "Welding demand is driven by construction, manufacturing, and infrastructure development. Hydropower projects, bridge construction, and building construction in Nepal all require welders. Skilled welders with multiple technique certifications earn more. Specialized welding (TIG for aluminum/stainless, pipe welding) commands premium rates. Foreign employment for skilled welders (Gulf, Australia, Canada) is an option with good pay.",
		Tags: []string{"skilled trades", "fabrication", "metal work", "hands-on"},
		Resources: []resourceSeed{
			{Title: "CTEVT Welding Programs", URL: "https://www.ctevt.org.np", Description: "Vocational training in welding and fabrication in Nepal"},
			{Title: "Merojob Welding Jobs", URL: "https://www.merojob.com", Description: "Find welder jobs in Nepal"},
			{Title: "Welding techniques for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=welding+for+beginners+nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn basic welding and metal work skills", Description: "Start with basic welding techniques — arc welding (stick welding) is the most common and accessible. Practice on scrap metal — learn to strike an arc, maintain a consistent arc length, and create a sound weld bead. Learn about safety: welding helmet with proper shade, leather gloves, fire-resistant clothing, ventilation, and fire safety. Welding is dangerous without proper safety. Learn metal cutting with grinders and cutting torches. Understand different metals and their welding properties (mild steel, stainless steel, aluminum). Practice daily — welding skill comes from hours of practice. Your early welds will be ugly — that is normal. Keep practicing until your welds are consistent and strong.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Arc welding basics for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=arc+welding+basics+for+beginners"},
				{Title: "Welding safety equipment and procedures", URL: "https://www.youtube.com/results?search_query=welding+safety+gear+guide"},
				{Title: "Different types of welding explained", URL: "https://www.youtube.com/results?search_query=types+of+welding+explained+stick+mig+tig"},
			}},
			{StepNumber: 2, Title: "Get formal welding training and certification", Description: "Enroll in a CTEVT welding and fabrication program. Get certified in specific welding techniques: SMAW (stick), GMAW (MIG), GTAW (TIG), and gas welding. Certification demonstrates competence and increases employability. Formal training teaches you: proper joint preparation, welding positions (flat, horizontal, vertical, overhead), weld inspection, and defect identification. Learn to read welding symbols on blueprints. TIG welding is the most skilled technique — used for aluminum, stainless steel, and precision work. MIG welding is faster and used for production work. A certified welder with multiple technique qualifications can command premium wages.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "CTEVT - Welding and Fabrication Program", URL: "https://www.ctevt.org.np"},
				{Title: "TIG welding for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=tig+welding+for+beginners+guide"},
				{Title: "MIG welding techniques and settings", URL: "https://www.youtube.com/results?search_query=mig+welding+techniques+for+beginners"},
			}},
			{StepNumber: 3, Title: "Work as a welder on projects", Description: "Start working at a fabrication shop, construction site, or manufacturing facility. Fabrication shops produce gates, railings, windows, furniture, and structural components. Construction sites need welders for steel structures, rebar, and embedded items. Manufacturing facilities need welders for production and maintenance. Each environment teaches different skills. Learn to work efficiently — set up quickly, weld cleanly, minimize grinding. Speed with quality is the mark of an experienced welder. Build a reputation for strong, clean welds. Inspect your own work critically. A weld that fails can cause injury, death, or costly damage. Quality and integrity are non-negotiable in welding.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Find welder jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Fabrication shop welding tips (YouTube)", URL: "https://www.youtube.com/results?search_query=fabrication+welding+tips+and+tricks"},
				{Title: "Weld inspection and quality control", URL: "https://www.youtube.com/results?search_query=weld+inspection+methods+quality"},
			}},
			{StepNumber: 4, Title: "Master advanced welding techniques", Description: "Learn advanced techniques: pipe welding (for pipelines, plumbing, boilers), structural welding (heavy steel for bridges and buildings), aluminum welding (TIG), stainless steel welding, and underwater welding (specialized, high pay). Get certified in advanced techniques: 6G pipe welding certification is the gold standard for pipe welders. Learn about welding metallurgy — how heat affects different metals and how to control distortion, cracking, and other weld defects. Learn about CNC plasma cutting and automated welding systems. Advanced welders who can handle complex jobs earn significantly more. Continuous skill development keeps you competitive.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Pipe welding - 6G certification guide (YouTube)", URL: "https://www.youtube.com/results?search_query=6g+pipe+welding+certification+guide"},
				{Title: "Structural welding for bridges and buildings", URL: "https://www.youtube.com/results?search_query=structural+welding+techniques+nepal"},
				{Title: "Aluminum TIG welding tips", URL: "https://www.youtube.com/results?search_query=aluminum+tig+welding+tips+and+tricks"},
			}},
			{StepNumber: 5, Title: "Start your own fabrication business", Description: "Start a welding and fabrication business: metal gates, railings, window grills, furniture, structural fabrication, and repair work. Investment needed: welding machines (arc, MIG, TIG), cutting equipment (grinders, plasma cutter), workspace, and materials. Target customers: homeowners (gates, railings), contractors (structural fabrication), businesses (shelving, racks, fixtures), and vehicle owners (repair work). Build relationships with construction companies and contractors for regular fabrication contracts. Quality work and reliable delivery build a strong reputation. Offer custom fabrication — unique designs that mass producers cannot offer. A skilled welder with business sense can build a successful enterprise.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Starting a welding business in Nepal", URL: "https://www.youtube.com/results?search_query=start+welding+business+nepal"},
				{Title: "Metal gate and railing fabrication business", URL: "https://www.youtube.com/results?search_query=metal+gate+fabrication+business+nepal"},
				{Title: "Welding business equipment and investment guide", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 6, Title: "Become a master welder or welding instructor", Description: "Master welders are highly respected — they can weld any material in any position, handle complex projects, and teach others. Some become welding inspectors (certified welding inspector - CWI) who certify welds for critical projects. Others become instructors at CTEVT institutes, training the next generation of welders. Welding skills also enable foreign employment — skilled welders are in demand in Australia, Canada, Gulf countries, and Europe. Welding is a trade that provides worldwide opportunities. A master welder's skills are valued anywhere metal is fabricated. Your welds hold together the structures that shape our world.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "AWS Certified Welding Inspector (CWI) path", URL: "https://www.youtube.com/results?search_query=aws+cwi+certification+guide"},
				{Title: "Welding instructor career at CTEVT", URL: "https://www.ctevt.org.np"},
				{Title: "Foreign employment for Nepali welders", URL: "https://www.youtube.com/results?search_query=nepali+welders+foreign+employment+australia"},
			}},
		},
	}
}
func beautyParlor() careerSeed {
	return careerSeed{
		CategoryName: "Service & Retail", CategorySlug: "service-retail", CategoryIcon: "💅",
		Title: "Beauty Parlor Owner/Beautician", Slug: "beauty-parlor",
		Summary: "Beauticians provide skincare, makeup, haircare, and nail services to clients. They work in beauty parlors, salons, spas, or run their own businesses.",
		Description: "A beautician provides beauty services: facials, makeup, hair cutting and styling, manicure/pedicure, threading, waxing, and skincare treatments. Beauty parlors and salons are everywhere in Nepal — from small neighborhood shops to high-end urban salons. The beauty industry in Nepal has grown tremendously with increasing disposable income and fashion consciousness. Beauticians can work as employees or start their own parlor. Good beauticians build loyal clients who return regularly. The career requires good customer service skills, creativity, and knowledge of beauty products and techniques. Beauty parlors often expand into additional services like bridal makeup (a huge market in Nepal), spa services, and beauty product sales.",
		DailyTasks: []string{"Provide facials, skincare treatments, and makeup services", "Cut, style, and color hair", "Perform manicure and pedicure", "Threading, waxing, and other hair removal services", "Advise clients on skincare and beauty products", "Manage appointments and client records", "Maintain cleanliness and hygiene in the parlor"},
		Skills: []string{"Skincare knowledge and facial techniques", "Makeup application (daily, bridal, fashion)", "Hair cutting, styling, and coloring", "Manicure, pedicure, and nail art", "Customer service and communication", "Business management for parlor owners", "Hygiene and sanitation standards", "Product knowledge and retail skills"},
		SalaryMin: 150000, SalaryMax: 600000, Difficulty: 2, FutureProof: 50,
		EducationReq: "SEE pass. Beauty culture training from institutes like LTW (Lalitpur Technical Workshop) or Impa Beauty. Diploma in Cosmetology. Apprenticeship in established beauty parlor. Product brand certifications (Lakme, VLCC, etc.). Advanced training in bridal makeup and hair styling.",
		Outlook: "Beauty industry is growing with rising beauty consciousness. Bridal makeup is a huge market in Nepal. Urban salons and high-end services are expanding. Competition is high but quality beauticians with good customer service build loyal clientele. The industry offers both employment and self-employment opportunities. Product sales add income.",
		Tags: []string{"service", "beauty", "salon", "entrepreneurship"},
		Resources: []resourceSeed{
			{Title: "CTEVT Cosmetology Programs", URL: "https://www.ctevt.org.np", Description: "Beauty and cosmetology training in Nepal"},
			{Title: "Merojob Beauty Jobs", URL: "https://www.merojob.com", Description: "Find beautician jobs in Nepal"},
			{Title: "Beauty parlor business tips (YouTube)", URL: "https://www.youtube.com/results?search_query=beauty+parlor+business+nepal+tips"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn basic beauty services and techniques", Description: "Start by learning the basics of beauty services: facial techniques, basic makeup, hair care, and nail care. Practice on family and friends. Watch tutorials and learn from experienced beauticians. Understand skin types, hair types, and basic product knowledge. Learn hygiene and sanitation standards — in beauty services, cleanliness is non-negotiable. Develop good customer service skills — being friendly, professional, and listening to clients' needs. If you enjoy making people feel good about themselves and have an eye for detail, beauty services could be your path. Practice every skill until you can do it well and efficiently.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Facial techniques for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=facial+techniques+for+beginners+nepal"},
				{Title: "Basic makeup tutorial for beginners", URL: "https://www.youtube.com/results?search_query=basic+makeup+tutorial+for+beginners+nepali"},
				{Title: "Hygiene standards for beauty parlors", URL: "https://www.youtube.com/results?search_query=beauty+parlor+hygiene+standards"},
			}},
			{StepNumber: 2, Title: "Get formal beauty training", Description: "Enroll in a beauty training program at a recognized institute. Complete a diploma in Cosmetology or Beauty Culture. Training covers: skincare science, makeup artistry, hair cutting and styling, nail technology, and salon management. Get certified in specific skills: bridal makeup (huge demand in Nepal), hair coloring and highlights, advanced facial techniques (chemical peels, microdermabrasion), and nail art. Product brand certifications (Lakme, VLCC, Oriflame) add credibility and open job opportunities. A trained beautician with certifications is more employable and can charge more for services. Continuous learning is important — beauty trends change constantly.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "CTEVT - Cosmetology Diploma", URL: "https://www.ctevt.org.np"},
				{Title: "Bridal makeup training Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=bridal+makeup+training+nepal"},
				{Title: "Hair coloring and highlights techniques", URL: "https://www.youtube.com/results?search_query=hair+coloring+techniques+for+salon"},
			}},
			{StepNumber: 3, Title: "Work in a beauty parlor to gain experience", Description: "Start working at an established beauty parlor or salon. You will learn real-world skills: handling different client types, managing appointments, retailing products, and working efficiently. Experienced beauticians and salon owners can teach you advanced techniques and business skills. Build your own clientele by being friendly, professional, and skilled. Clients who like you will ask for you specifically — this is the mark of a successful beautician. Learn about salon hygiene standards, equipment maintenance, and product inventory. Work experience is essential before starting your own parlor. Learn the business side too — pricing, costs, customer management.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Find parlor jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Salon customer service tips (YouTube)", URL: "https://www.youtube.com/results?search_query=salon+customer+service+tips"},
				{Title: "Building a beauty client base", URL: "https://www.youtube.com/results?search_query=build+beauty+client+base+salon"},
			}},
			{StepNumber: 4, Title: "Specialize in high-demand services", Description: "Develop expertise in high-value services: bridal makeup (packages for wedding parties — a lucrative niche), advanced skincare (chemical peels, microdermabrasion, LED therapy), hair extensions and weaves, nail art and gel nails, or Spa services (massage, body treatments). Bridal makeup is especially profitable in Nepal — weddings are elaborate events with multiple makeup needs (bride, family, bridesmaids). Specialization allows you to charge premium prices. Build a portfolio of your best work — before/after photos, bridal looks, creative styles. Social media is essential for beauty professionals. Instagram and Facebook showcase your work and attract clients.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Bridal makeup portfolio building (YouTube)", URL: "https://www.youtube.com/results?search_query=bridal+makeup+portfolio+tips"},
				{Title: "Advanced skincare treatments training", URL: "https://www.youtube.com/results?search_query=advanced+skincare+treatments+training"},
				{Title: "Instagram marketing for beauticians", URL: "https://www.youtube.com/results?search_query=instagram+for+beauticians+marketing"},
			}},
			{StepNumber: 5, Title: "Open your own beauty parlor", Description: "Starting your own parlor requires: a good location (residential area with good foot traffic, or commercial area), equipment (beauty chairs, facial beds, hair washing stations, product inventory), a small team if needed (hire one or two beauticians if demand is high), business registration, and a clean, attractive space. Start small and expand as your clientele grows. Offer a range of services but focus on what you do best. Build relationships with bridal shops and event planners for referrals. A well-run parlor provides stable income and the satisfaction of being your own boss. The beauty industry in Nepal has room for quality-focused businesses.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Starting a beauty parlor in Nepal", URL: "https://www.youtube.com/results?search_query=start+beauty+parlor+nepal+guide"},
				{Title: "Beauty parlor equipment and setup cost", URL: "https://www.merojob.com"},
				{Title: "Salon business management tips", URL: "https://www.youtube.com/results?search_query=salon+business+management+nepal"},
			}},
			{StepNumber: 6, Title: "Expand your beauty business or teach", Description: "Grow your parlor: add more services (spa, massage, bridal studio), open additional locations, introduce beauty product retail (selling the products you use), or become a distributor for beauty brands. Some experienced beauticians become trainers at beauty institutes, teaching the next generation of professionals. The beauty industry values skill and reputation. A beautician who builds a strong brand can create a lasting business. Your work helps people look their best and feel confident. For brides especially, a beautician's work is part of one of the most important days of their lives.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Expanding your salon business (YouTube)", URL: "https://www.youtube.com/results?search_query=salon+business+expansion+strategies"},
				{Title: "Beauty product retail business", URL: "https://www.youtube.com/results?search_query=beauty+product+retail+business+nepal"},
				{Title: "Beauty training and teaching career", URL: "https://www.ctevt.org.np"},
			}},
		},
	}
}

func eventManager() careerSeed {
	return careerSeed{
		CategoryName: "Service & Retail", CategorySlug: "service-retail", CategoryIcon: "🎉",
		Title: "Event Manager", Slug: "event-manager",
		Summary: "Event managers plan, organize, and execute events — weddings, conferences, festivals, corporate events, and social gatherings in Nepal.",
		Description: "An event manager plans and executes events of all types: weddings (a huge market in Nepal), corporate events (conferences, product launches, team building), social events (birthdays, anniversaries, parties), cultural festivals, music concerts, and sporting events. Event management in Nepal has grown significantly with the wedding industry, corporate sector expansion, and increasing demand for professionally organized events. Event managers handle everything — venue selection, decoration, catering, entertainment, scheduling, budgeting, and client coordination. The job requires excellent organizational skills, creativity, and the ability to manage multiple things simultaneously. Successful event managers build reputations for creating memorable experiences. Events in Nepal, especially weddings, are elaborate affairs that require professional coordination.",
		DailyTasks: []string{"Meet with clients to understand their event requirements", "Plan event details — venue, decoration, catering, entertainment, schedule", "Coordinate with vendors — caterers, decorators, photographers, musicians", "Manage event budgets and negotiate with suppliers", "Create event timelines and run sheets", "Supervise event setup and execution on the day", "Handle problems and last-minute changes during events"},
		Skills: []string{"Event planning and project management", "Vendor negotiation and relationship management", "Budgeting and financial management", "Creativity and design sense for event decoration", "Problem-solving under pressure", "Communication with clients, vendors, and team", "Time management and ability to multitask", "Knowledge of Nepali event traditions and customs"},
		SalaryMin: 240000, SalaryMax: 1200000, Difficulty: 3, FutureProof: 60,
		EducationReq: "Bachelor's in Business Management, Marketing, or Hospitality preferred. Event management certification from institutes like NCHMCT or event management training programs. Practical experience through internships or assistant roles is essential. Knowledge of wedding traditions important.",
		Outlook: "Event management is growing with Nepal's wedding industry, corporate events, and tourism events. Wedding planning is the largest segment. Corporate conferences and product launches are increasing. Festival and concert events are growing. Competition exists but quality event managers with good reputations have steady business. The industry is relationship-driven — success depends on vendor network and client referrals.",
		Tags: []string{"service", "events", "weddings", "entrepreneurship"},
		Resources: []resourceSeed{
			{Title: "Event Management Association Nepal", URL: "https://www.eman.org.np", Description: "Professional body for event managers in Nepal"},
			{Title: "Merojob Event Jobs", URL: "https://www.merojob.com", Description: "Find event management jobs in Nepal"},
			{Title: "Wedding planning tips Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=wedding+planning+nepal+guide"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Develop organizational and planning skills", Description: "Event management requires excellent organizational skills. Start by helping organize events for family, friends, or community groups — weddings, parties, school programs. Learn to create checklists, timelines, and budgets. Practice coordinating with multiple people — vendors, helpers, family members. Develop problem-solving skills — things always go wrong at events and you need to fix them quickly. Study event management principles online. Learn about different event types and their requirements. Observe established event managers at work. If you love planning, organizing, and seeing your plans come together, event management could be your career.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Event planning basics for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=event+planning+basics+for+beginners"},
				{Title: "Event management checklists and templates", URL: "https://www.youtube.com/results?search_query=event+management+checklist+guide"},
				{Title: "Nepali wedding traditions and rituals explained", URL: "https://www.youtube.com/results?search_query=nepali+wedding+traditions+guide"},
			}},
			{StepNumber: 2, Title: "Get education and training in event management", Description: "Pursue a degree in Business Management, Marketing, or Hospitality, or take an event management certification course. Learn about: event design, budgeting, vendor management, risk management, marketing, and client management. Study different event types — weddings, corporate events, festivals, exhibitions, conferences. Nepali wedding traditions are complex with many rituals — knowledge of these is essential for wedding planning. Learn about event decoration trends — Nepali weddings feature elaborate decoration (mandap, flowers, lighting). Learn to use event management software for planning and coordination. Formal training gives you credibility and a systematic approach.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "NCHMCT Event Management Programs", URL: "https://www.nchmct.edu.np"},
				{Title: "Nepali wedding decoration trends (YouTube)", URL: "https://www.youtube.com/results?search_query=nepali+wedding+decoration+ideas"},
				{Title: "Event management software for planners", URL: "https://www.youtube.com/results?search_query=event+management+software+tools"},
			}},
			{StepNumber: 3, Title: "Work as an assistant to an experienced event manager", Description: "Join an event management company or work as an assistant to an established event planner. Learn the practical aspects: how to negotiate with vendors, how to create and manage event budgets, how to handle difficult clients, how to manage event-day logistics. Learn the vendor network — caterers, decorators, photographers, videographers, musicians, tent suppliers, and transportation providers. Good vendor relationships are essential for event managers. Learn from mistakes — events always have unexpected challenges. Each event teaches new lessons. Build your reputation by being reliable, professional, and calm under pressure.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Event management internships Nepal", URL: "https://www.merojob.com"},
				{Title: "Vendor management for event planners (YouTube)", URL: "https://www.youtube.com/results?search_query=vendor+management+for+event+planners"},
				{Title: "Handling event crises and problems", URL: "https://www.youtube.com/results?search_query=event+crisis+management+tips"},
			}},
			{StepNumber: 4, Title: "Build a portfolio and specialize", Description: "Document every event you work on — photographs, videos, client testimonials. Build a portfolio showcasing your best work. Create a website or social media presence (Instagram is essential for visual events like weddings). Specialize in a type of event: wedding planning (the biggest market in Nepal), corporate events (conferences, retreats, product launches), cultural festivals (Indra Jatra, Dashain events, theme parties), or destination events (events in Pokhara, Chitwan, or trekking destinations). Specialization allows you to build deep expertise and reputation. Wedding planners who understand all the rituals and traditions are particularly valued in Nepal.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Building an event planning portfolio (YouTube)", URL: "https://www.youtube.com/results?search_query=event+planning+portfolio+tips"},
				{Title: "Destination event planning in Nepal", URL: "https://www.youtube.com/results?search_query=destination+wedding+nepal+planning"},
				{Title: "Corporate event management guide", URL: "https://www.youtube.com/results?search_query=corporate+event+planning+nepal"},
			}},
			{StepNumber: 5, Title: "Start your own event management company", Description: "Starting an event management company requires: business registration, a network of reliable vendors, initial marketing (portfolio, website, social media), and possibly a small team. Focus on quality and reliability — word of mouth is the most powerful marketing. Build packages for different event types and budgets. Develop a unique selling proposition — what makes your event management different? It could be your decoration style, attention to detail, vendor relationships, or pricing. Build relationships with venues, hotels, and restaurants for referral business. Event management can be seasonal — weddings peak in certain months. Manage finances accordingly.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Starting an event management business in Nepal", URL: "https://www.youtube.com/results?search_query=start+event+management+company+nepal"},
				{Title: "Event business marketing strategies", URL: "https://www.youtube.com/results?search_query=event+business+marketing+nepal"},
				{Title: "Wedding package pricing guide for planners", URL: "https://www.youtube.com/results?search_query=wedding+package+pricing+for+planners"},
			}},
			{StepNumber: 6, Title: "Establish yourself as a top event organizer", Description: "Top event managers are known for creating exceptional experiences. They have a strong brand, loyal clients, and a network of the best vendors. Some expand into related businesses — decoration rental, event equipment rental, catering, or venue management. Others train new event managers or consult on large-scale events like festivals, concerts, and international conferences. Events bring people together to celebrate, do business, learn, and connect. An event manager creates the space for these experiences. Every successful event you organize is a memory that people will cherish. Your work makes celebrations happen.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Scaling an event management business (YouTube)", URL: "https://www.youtube.com/results?search_query=scale+event+management+business"},
				{Title: "Event equipment rental business Nepal", URL: "https://www.merojob.com"},
				{Title: "Large-scale festival and concert management", URL: "https://www.youtube.com/results?search_query=festival+event+management+nepal"},
			}},
		},
	}
}
func sportsCoach() careerSeed {
	return careerSeed{
		CategoryName: "Service & Retail", CategorySlug: "service-retail", CategoryIcon: "🏅",
		Title: "Sports Coach", Slug: "sports-coach",
		Summary: "Sports coaches train athletes and teams in various sports, develop skills, design training programs, and guide athletes to reach their full potential.",
		Description: "A sports coach trains athletes and teams in specific sports — football (the most popular in Nepal), cricket, basketball, volleyball, martial arts, athletics, badminton, table tennis, and traditional Nepali sports. Coaches teach techniques, develop fitness, design training programs, plan game strategies, and motivate athletes. Nepal has growing interest in sports with domestic leagues (Nepal Super League for football, Prime Minister Cup for cricket), school and college sports programs, and international competitions. Coaches work in schools, sports academies, clubs, and community programs. Good coaches understand both the technical aspects of their sport and how to motivate and develop athletes. Sports coaching is rewarding — you help athletes achieve their potential and contribute to Nepal's sports development.",
		DailyTasks: []string{"Plan and conduct training sessions for athletes", "Demonstrate techniques and correct athletes' form", "Develop fitness and conditioning programs", "Analyze game footage and plan match strategies", "Provide feedback and motivation to athletes", "Monitor athlete progress and adjust training", "Coordinate with parents, school officials, or club management"},
		Skills: []string{"Deep knowledge of the specific sport (rules, techniques, strategies)", "Training program design for different age/ability levels", "Communication and motivation skills", "First aid and injury prevention knowledge", "Fitness and conditioning expertise", "Patience and ability to work with different personalities", "Game analysis and strategic thinking", "Leadership and team building"},
		SalaryMin: 180000, SalaryMax: 900000, Difficulty: 3, FutureProof: 50,
		EducationReq: "Bachelor's in Physical Education or Sports Science preferred. Coaching certification from Nepal Sports Council or international bodies (FIFA, ICC, etc.). First aid and CPR certification. Experience as a player in the sport is highly valued. Continuous learning through coaching clinics and workshops.",
		Outlook: "Sports development in Nepal is growing with professional leagues, school sports programs, and government investment. Football and cricket coaching are in highest demand. Qualified coaches are needed at schools, academies, and clubs. Sports science and professional coaching standards are improving. Coaching at the professional level (national team, franchise leagues) offers higher pay and prestige.",
		Tags: []string{"service", "sports", "coaching", "fitness"},
		Resources: []resourceSeed{
			{Title: "Nepal Sports Council", URL: "https://www.nocnepal.org.np", Description: "National sports authority and coaching certification"},
			{Title: "ANFA (All Nepal Football Association)", URL: "https://www.anfa.org.np", Description: "Football coaching and development in Nepal"},
			{Title: "Cricket Association of Nepal (CAN)", URL: "https://www.cricketnepal.org.np", Description: "Cricket coaching and certification in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Master the sport you want to coach", Description: "You cannot coach a sport you do not understand deeply. Play the sport at a competitive level. Understand the rules, techniques, strategies, and nuances of the game. Study the sport — watch professional matches, read coaching books, analyze game strategies. Learn different coaching philosophies and systems used internationally. Develop your understanding of how the sport is played at different levels — beginner, intermediate, advanced. Experience as a player gives you credibility and practical insights that coaching courses cannot teach. The best coaches were not always the best players, but they deeply understand the sport and can communicate that understanding to others.",
			Duration: "2-5 years", Links: []roadmapLink{
				{Title: "Football coaching fundamentals (YouTube)", URL: "https://www.youtube.com/results?search_query=football+coaching+basics+nepal"},
				{Title: "Cricket coaching techniques and drills", URL: "https://www.youtube.com/results?search_query=cricket+coaching+drills+for+beginners"},
				{Title: "Sports science and coaching principles", URL: "https://www.youtube.com/results?search_query=sports+science+coaching+principles"},
			}},
			{StepNumber: 2, Title: "Get coaching education and certification", Description: "Pursue a degree in Physical Education or Sports Science. Get coaching certification from Nepal Sports Council or international sports bodies. FIFA offers football coaching licenses (C, B, A license). The ICC offers cricket coaching certification. ANFA and CAN conduct regular coaching courses in Nepal. Learn about: sports physiology, training methodology, nutrition for athletes, injury prevention and management, sports psychology, and coaching pedagogy. Attend coaching clinics and workshops. Good coaches are lifelong learners — the science of sports training evolves constantly. Certification demonstrates professional standards and commitment to quality coaching.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Nepal Sports Council - Coaching Certification", URL: "https://www.nocnepal.org.np"},
				{Title: "FIFA coaching license program (YouTube)", URL: "https://www.youtube.com/results?search_query=fifa+coaching+license+program"},
				{Title: "Sports psychology for coaches", URL: "https://www.youtube.com/results?search_query=sports+psychology+coaching+tips"},
			}},
			{StepNumber: 3, Title: "Start coaching at school or community level", Description: "Begin coaching at a school, community club, or local sports academy. Start with young or beginner athletes — learning to teach fundamentals to beginners develops your coaching skills. Design training sessions, plan season programs, and manage teams. Learn to work with different age groups and skill levels. Coaching children requires patience and the ability to make training fun. Coaching advanced athletes requires tactical knowledge and intensity. Build relationships with athletes — good coaches know when to push and when to support. Each practice session and game is a learning opportunity for you as a coach.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Find sports coaching jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Youth sports coaching tips (YouTube)", URL: "https://www.youtube.com/results?search_query=youth+sports+coaching+tips"},
				{Title: "Planning a training session guide", URL: "https://www.youtube.com/results?search_query=training+session+planning+for+coaches"},
			}},
			{StepNumber: 4, Title: "Specialize in coaching methodology or athlete development", Description: "Develop expertise in: talent identification and development (finding and nurturing young talent), strength and conditioning coaching (fitness training for athletes), sports psychology (mental preparation, motivation, focus), technical skills coaching (specialist skills like bowling in cricket, goalkeeping in football), or tactical analysis (game strategy, opponent analysis). Specialization makes you more valuable. A strength and conditioning coach is needed by all sports teams. A specialist goalkeeping coach or batting coach provides focused expertise that general coaches cannot. Consider taking specialized certification in your chosen area.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Strength and conditioning coach certification", URL: "https://www.youtube.com/results?search_query=strength+and+conditioning+certification"},
				{Title: "Talent identification in sports (YouTube)", URL: "https://www.youtube.com/results?search_query=talent+identification+in+sports+nepal"},
				{Title: "Game analysis and video coaching tools", URL: "https://www.youtube.com/results?search_query=game+analysis+software+for+coaches"},
			}},
			{StepNumber: 5, Title: "Coach at higher levels or open your own academy", Description: "Progress to coaching at higher levels — district teams, provincial teams, club franchises in professional leagues (NSL, PM Cup), or national team age-group levels. Higher-level coaching requires deeper tactical knowledge, man-management skills, and the ability to perform under pressure (results matter). Alternatively, open your own sports academy or coaching center. An academy provides training to athletes who pay for coaching. Running an academy requires business skills — marketing, facilities management, fee collection, and staff management. Many successful coaches in Nepal run their own academies while also coaching teams.",
			Duration: "3-5 years", Links: []roadmapLink{
				{Title: "Nepal Super League coaching opportunities", URL: "https://www.youtube.com/results?search_query=nepal+super+league+coach+role"},
				{Title: "Starting a sports academy in Nepal", URL: "https://www.youtube.com/results?search_query=start+sports+academy+nepal"},
				{Title: "Athlete management for high-level coaches", URL: "https://www.youtube.com/results?search_query=managing+elite+athletes+coaching"},
			}},
			{StepNumber: 6, Title: "Become a national coach or sports development leader", Description: "Top coaches become national team coaches — the highest level in Nepali sports. National coaches prepare teams for international competitions (SAFF Championship, ACC tournaments, Asian Games, Olympics). Some coach professional teams abroad. Others move into sports administration — managing national sports federations, developing coaching education programs, or leading sports policy. Sports coaching in Nepal is a growing profession. As sports develop professionally, the demand for qualified, dedicated coaches increases. Your coaching helps athletes achieve their dreams and puts Nepal on the international sports map.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Nepal national team coach career path", URL: "https://www.nocnepal.org.np"},
				{Title: "Sports administration and management (YouTube)", URL: "https://www.youtube.com/results?search_query=sports+administration+nepal"},
				{Title: "International coaching opportunities", URL: "https://www.youtube.com/results?search_query=international+sports+coaching+career"},
			}},
		},
	}
}

func gymTrainer() careerSeed {
	return careerSeed{
		CategoryName: "Service & Retail", CategorySlug: "service-retail", CategoryIcon: "💪",
		Title: "Gym Trainer (Fitness Instructor)", Slug: "gym-trainer",
		Summary: "Gym trainers help clients achieve fitness goals through exercise programs, proper technique, nutrition guidance, and motivation in gyms and fitness centers.",
		Description: "A gym trainer (fitness instructor) designs and supervises exercise programs for clients in gyms and fitness centers. They demonstrate proper exercise technique, create personalized workout plans, provide nutrition guidance, and motivate clients to achieve their fitness goals. Nepal's fitness industry has grown rapidly — gyms are everywhere in urban areas, from small neighborhood gyms to large fitness chains. Trainers work with diverse clients — from beginners wanting to lose weight to athletes wanting to build strength. The career requires fitness knowledge, good communication skills, and genuine interest in helping people improve their health. Good trainers build loyal client bases and can earn well through personal training sessions.",
		DailyTasks: []string{"Design personalized workout programs for clients", "Demonstrate exercise techniques and correct form", "Supervise clients during workouts for safety", "Provide nutrition and lifestyle advice", "Lead group fitness classes (aerobics, yoga, HIIT)", "Maintain gym equipment and cleanliness", "Track client progress and adjust programs accordingly"},
		Skills: []string{"Exercise science knowledge (anatomy, physiology, biomechanics)", "Program design for different goals (weight loss, muscle gain, fitness)", "Exercise technique instruction and spotting", "Nutrition basics and dietary guidance", "Motivation and coaching psychology", "Group fitness class instruction", "Customer service and interpersonal skills", "First aid and CPR certification"},
		SalaryMin: 180000, SalaryMax: 700000, Difficulty: 2, FutureProof: 55,
		EducationReq: "SEE/+2 pass. Certification in Fitness Training from recognized institutes (ISSA, ACE, NASM, or local programs). CPR and First Aid certification. Nutrition course helpful. Practical experience with weight training essential. Continuous education in fitness trends.",
		Outlook: "Fitness industry in Nepal is growing with health awareness and rising disposable income. More gyms opening in urban and semi-urban areas. Certified trainers with good people skills are in demand. Personal training offers higher earnings than general floor instruction. Group fitness classes (Zumba, yoga, HIIT) are growing. Online fitness coaching is an emerging opportunity. The industry is competitive — building a client base takes time.",
		Tags: []string{"service", "fitness", "health", "training"},
		Resources: []resourceSeed{
			{Title: "Fitness trainer certification Nepal", URL: "https://www.youtube.com/results?search_query=fitness+trainer+certification+nepal"},
			{Title: "Merojob Fitness Jobs", URL: "https://www.merojob.com", Description: "Find gym trainer jobs in Nepal"},
			{Title: "Exercise science fundamentals (YouTube)", URL: "https://www.youtube.com/results?search_query=exercise+science+for+personal+trainers"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Build your own fitness and knowledge base", Description: "A gym trainer should practice what they preach. Start your own fitness journey — learn to train effectively, understand nutrition, and achieve fitness results. Study exercise science: anatomy (muscle groups, bones), physiology (how the body responds to exercise), and basic nutrition (macronutrients, calories, meal timing). Learn proper exercise technique for all major exercises — squat, deadlift, bench press, pull-ups, rows, lunges, and isolation exercises. Practice until you can demonstrate perfect form. Read fitness books, follow reputable fitness experts online, and study for personal trainer certification. Your own transformation gives you credibility with clients.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Anatomy and exercise science basics (YouTube)", URL: "https://www.youtube.com/results?search_query=anatomy+for+fitness+trainers+basics"},
				{Title: "Proper exercise form guide for major lifts", URL: "https://www.youtube.com/results?search_query=proper+squat+deadlift+bench+form"},
				{Title: "Nutrition basics for fitness", URL: "https://www.youtube.com/results?search_query=nutrition+basics+for+fitness+beginners"},
			}},
			{StepNumber: 2, Title: "Get certified as a personal trainer", Description: "Obtain a recognized personal trainer certification. International certifications (ISSA, ACE, NASM, ACSM) are respected globally. Local certification programs are available in Nepal through fitness institutes. Certification covers: client assessment, program design, exercise technique, nutrition basics, special populations (elderly, pregnant, injured), and business skills for trainers. Get CPR and First Aid certified — essential for gym safety. A certified trainer is more credible and employable. Some gyms require certification before hiring. Certification also provides liability insurance options. Online certification courses offer flexibility to study while continuing other work.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "ISSA personal trainer certification overview", URL: "https://www.youtube.com/results?search_query=issa+personal+trainer+certification"},
				{Title: "CPR and First Aid certification Nepal", URL: "https://www.youtube.com/results?search_query=cpr+first+aid+certification+nepal"},
				{Title: "ACE personal trainer exam prep (YouTube)", URL: "https://www.youtube.com/results?search_query=ace+personal+trainer+exam+prep"},
			}},
			{StepNumber: 3, Title: "Work as a gym instructor", Description: "Start working at a gym as a floor instructor or junior trainer. Help members with equipment, demonstrate exercises, and provide basic guidance. Learn the gym environment — how different gyms operate, member management, and sales. Build your experience training different types of clients — beginners, seniors, athletes, weight loss clients. Each client type requires different approaches. Develop your communication and motivation skills. A good trainer can make clients feel comfortable, confident, and motivated. Build relationships with gym members — they may become your personal training clients. Learn the business side — how gyms retain members, sell training packages, and manage schedules.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Find gym trainer jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Client communication and motivation tips (YouTube)", URL: "https://www.youtube.com/results?search_query=client+motivation+for+personal+trainers"},
				{Title: "Gym floor etiquette and member service", URL: "https://www.youtube.com/results?search_query=gym+etiquette+guide+for+trainers"},
			}},
			{StepNumber: 4, Title: "Build your personal training client base", Description: "Offer personal training sessions to clients who want one-on-one attention. Start with a few clients and build your reputation. Set competitive rates — personal training in Nepal ranges from 500-2000 NPR per session depending on your experience and location. Create client programs that deliver results — visible results are the best marketing. Get testimonials and before/after photos (with permission). Ask satisfied clients for referrals. Develop specialized programs: weight loss, muscle building, postnatal fitness, sports performance, or senior fitness. Specialization helps you stand out and charge higher rates. Build an online presence — Instagram and Facebook to showcase client transformations.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Building a personal training business (YouTube)", URL: "https://www.youtube.com/results?search_query=build+personal+training+business+nepal"},
				{Title: "Social media marketing for fitness trainers", URL: "https://www.youtube.com/results?search_query=social+media+for+fitness+trainers"},
				{Title: "Pricing personal training sessions guide", URL: "https://www.youtube.com/results?search_query=how+to+price+personal+training"},
			}},
			{StepNumber: 5, Title: "Expand into group classes or specialized training", Description: "Group fitness classes (Zumba, HIIT, Spin, Yoga, Bootcamp) are popular and can be more lucrative than one-on-one training — you train 10-30 people at once. Get certified in specific group fitness formats. Develop your own training style or signature class. Offer specialized services: sports-specific training (football, cricket fitness), weight loss transformation programs (structured programs with nutrition coaching), online personal training (coaching clients remotely via apps and video calls), or corporate wellness programs (workplace fitness sessions). Each expansion increases your income potential and client reach.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Group fitness instructor certification (YouTube)", URL: "https://www.youtube.com/results?search_query=group+fitness+instructor+certification"},
				{Title: "Online personal training business model", URL: "https://www.youtube.com/results?search_query=online+personal+training+business"},
				{Title: "Corporate wellness program setup", URL: "https://www.youtube.com/results?search_query=corporate+wellness+fitness+program"},
			}},
			{StepNumber: 6, Title: "Open your own fitness studio or become a master trainer", Description: "Top trainers open their own fitness studios or boutique gyms — offering specialized training in a dedicated space. A studio requires less investment than a full gym but offers more control over your training environment. Become a master trainer — mentoring new trainers, conducting certification courses, or consulting for gyms on training programs. Some trainers become fitness educators, teaching at fitness institutes. Others become fitness influencers and content creators, earning through sponsorships and online programs. The fitness industry rewards those who combine knowledge, people skills, and business acumen. Your work helps people live healthier, stronger, and more confident lives.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a fitness studio in Nepal", URL: "https://www.youtube.com/results?search_query=start+fitness+studio+nepal"},
				{Title: "Master trainer - mentoring new fitness professionals", URL: "https://www.youtube.com/results?search_query=mentoring+fitness+trainers+career"},
				{Title: "Fitness influencer and content creator tips", URL: "https://www.youtube.com/results?search_query=fitness+influencer+nepal+guide"},
			}},
		},
	}
}
func supermarketManager() careerSeed {
	return careerSeed{
		CategoryName: "Service & Retail", CategorySlug: "service-retail", CategoryIcon: "🛒",
		Title: "Supermarket Manager", Slug: "supermarket-manager",
		Summary: "Supermarket managers oversee daily operations of retail stores — managing staff, inventory, customer service, sales, and profitability.",
		Description: "A supermarket manager runs the daily operations of a retail store. They manage staff (hiring, training, scheduling), oversee inventory (ordering, stocking, managing waste), ensure customer satisfaction, handle finances (sales, expenses, profitability), and maintain store standards (cleanliness, safety, product quality). Nepal's retail sector is growing with modern supermarkets and department stores in urban areas (Bhatbhateni, Bigmart, Saleways, Superstar, and many local chains). Supermarket managers need retail experience, leadership skills, and knowledge of retail operations. The role requires dealing with customers, suppliers, and staff on a daily basis. It is a fast-paced environment where no two days are the same. Good managers create efficient, profitable stores with satisfied customers and motivated staff.",
		DailyTasks: []string{"Open and close the store following procedures", "Supervise staff — assign tasks, monitor performance, provide training", "Check inventory levels and place orders with suppliers", "Arrange product displays and ensure shelves are stocked", "Handle customer complaints and ensure good service", "Review sales reports and manage store budgets", "Ensure store cleanliness, safety, and security"},
		Skills: []string{"Retail management and operations knowledge", "Inventory management and stock control", "Staff supervision, training, and scheduling", "Customer service orientation and complaint handling", "Sales analysis and financial management", "Vendor relationship and negotiation", "Computer skills (POS systems, MS Office)", "Problem-solving and decision-making under pressure"},
		SalaryMin: 300000, SalaryMax: 1200000, Difficulty: 3, FutureProof: 60,
		EducationReq: "Bachelor's in Business Management, Retail Management, or related field preferred. Experience in retail operations is essential — most managers start as sales associates. Training in inventory management, customer service, and retail software. MBA can help for senior retail management positions.",
		Outlook: "Modern retail is expanding in Nepal with new supermarkets and shopping malls. Experienced retail managers are in demand. The sector offers career growth from assistant manager to store manager to regional manager. E-commerce is growing but brick-and-mortar retail remains strong. Retail skills are transferable to other industries.",
		Tags: []string{"service", "retail", "management", "business"},
		Resources: []resourceSeed{
			{Title: "Merojob Retail Jobs", URL: "https://www.merojob.com", Description: "Find supermarket and retail management jobs in Nepal"},
			{Title: "Bhatbhateni Supermarket Careers", URL: "https://www.bhatbhateni.com.np", Description: "Nepal's largest retail chain - career opportunities"},
			{Title: "Retail management principles (YouTube)", URL: "https://www.youtube.com/results?search_query=retail+management+basics+for+beginners"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Start working in retail to learn the basics", Description: "Begin your retail career as a sales associate, cashier, or customer service representative. Learn the fundamentals: how a retail store operates, how to handle customers, how the POS system works, how products are displayed, and how inventory is managed. Understand the importance of customer service — in retail, the customer's experience determines success. Learn about different product categories and their handling requirements (fresh produce, packaged goods, frozen items, household products). Pay attention to how your store is managed — note what works and what does not. Build a reputation for being reliable, hardworking, and good with customers. Retail experience from the ground up is invaluable for future managers.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Find retail jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Customer service skills for retail (YouTube)", URL: "https://www.youtube.com/results?search_query=customer+service+skills+for+retail+nepal"},
				{Title: "Retail store operations basics", URL: "https://www.youtube.com/results?search_query=retail+store+operations+guide"},
			}},
			{StepNumber: 2, Title: "Get education in business or retail management", Description: "Pursue a degree in Business Management or Retail Management. Study: retail operations, supply chain management, inventory control, marketing, financial management, human resources, and customer relationship management. Learn about retail-specific concepts: gross margin, stock turnover, shrinkage, planograms, and category management. Understand the numbers that drive retail business. A good manager understands financial statements — profit and loss, balance sheets, cash flow. Take courses in retail software — POS systems, inventory management systems, and Microsoft Excel for analysis. Education combined with practical experience creates effective managers.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Business Management programs Nepal (Edusanjal)", URL: "https://www.edusanjal.com"},
				{Title: "Retail inventory management techniques (YouTube)", URL: "https://www.youtube.com/results?search_query=retail+inventory+management+techniques"},
				{Title: "Understanding retail financial metrics", URL: "https://www.youtube.com/results?search_query=retail+KPI+metrics+explained"},
			}},
			{StepNumber: 3, Title: "Become a department manager or assistant manager", Description: "Progress to a supervisory role — department manager (produce, grocery, dairy, etc.) or assistant store manager. Take on more responsibility: managing a team, ordering products, managing a department's budget, and ensuring department performance. Learn to handle staff issues — scheduling, performance feedback, training new employees. Handle customer complaints and difficult situations. Learn from the store manager — observe how they handle operations, manage people, and make decisions. Take initiative — suggest improvements, volunteer for special projects, show that you are ready for more responsibility. Your performance as assistant manager determines whether you will be promoted to store manager.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "From sales associate to manager - retail career path (YouTube)", URL: "https://www.youtube.com/results?search_query=retail+career+path+to+management"},
				{Title: "Staff management and leadership in retail", URL: "https://www.youtube.com/results?search_query=retail+staff+management+tips"},
				{Title: "Handling difficult customers in retail", URL: "https://www.youtube.com/results?search_query=handling+difficult+customers+retail+tips"},
			}},
			{StepNumber: 4, Title: "Take on store manager responsibilities", Description: "As store manager, you are responsible for everything: sales performance, profitability, inventory accuracy, staff management, customer satisfaction, and compliance with company policies and local regulations. Learn to analyze sales data and adjust strategies accordingly. Manage store budgets, labor costs, and operational expenses. Build a strong team — hire good people, train them well, and create a positive work environment. A successful store has low staff turnover, high customer satisfaction, and strong financial performance. Good store managers are visible on the floor, interact with customers, and lead by example. Your store's performance reflects your leadership.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Store manager responsibilities and skills (YouTube)", URL: "https://www.youtube.com/results?search_query=store+manager+role+and+responsibilities"},
				{Title: "Retail sales analysis and improvement strategies", URL: "https://www.youtube.com/results?search_query=retail+sales+analysis+methods"},
				{Title: "Building a strong retail team", URL: "https://www.youtube.com/results?search_query=build+strong+retail+team+management"},
			}},
			{StepNumber: 5, Title: "Specialize in retail operations or buying/merchandising", Description: "Senior retail professionals specialize in: buying and merchandising (selecting products, negotiating with suppliers, planning product ranges), operations management (overseeing multiple stores, logistics, supply chain), category management (managing specific product categories across stores), or retail marketing (promotions, loyalty programs, customer engagement). Buying and merchandising is a particularly important retail function — selecting the right products at the right price determines store success. Category managers are experts in their product categories — they understand market trends, customer preferences, and supplier dynamics. Specialization in these areas can lead to corporate retail positions at the head office level.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "Retail buying and merchandising career (YouTube)", URL: "https://www.youtube.com/results?search_query=retail+buying+merchandising+career"},
				{Title: "Category management in retail explained", URL: "https://www.youtube.com/results?search_query=category+management+retail+guide"},
				{Title: "Retail supply chain management basics", URL: "https://www.youtube.com/results?search_query=retail+supply+chain+management+nepal"},
			}},
			{StepNumber: 6, Title: "Become a regional manager or retail entrepreneur", Description: "Top retail managers become regional managers — overseeing multiple stores in a region, driving performance, developing store managers, and ensuring brand standards across locations. Some move into retail entrepreneurship — opening their own retail business. Starting a retail store requires capital, location, supplier relationships, and business planning. Nepal's retail sector offers opportunities for independent stores in neighborhoods and communities. Retail management skills are transferable to any business that sells products. Every retail manager plays a vital role in their community — providing jobs, serving customers, and contributing to the local economy.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Regional retail management career (YouTube)", URL: "https://www.youtube.com/results?search_query=regional+manager+retail+role"},
				{Title: "Starting a retail business in Nepal", URL: "https://www.youtube.com/results?search_query=start+retail+store+business+nepal"},
				{Title: "Retail industry trends and future in Nepal", URL: "https://www.merojob.com"},
			}},
		},
	}
}
func airlinePilot() careerSeed {
	return careerSeed{
		CategoryName: "Transport & Logistics", CategorySlug: "transport-logistics", CategoryIcon: "✈️",
		Title: "Airline Pilot", Slug: "airline-pilot",
		Summary: "Airline pilots fly aircraft for domestic and international airlines, transporting passengers and cargo safely and efficiently.",
		Description: "An airline pilot operates aircraft for commercial airlines, flying passengers or cargo on domestic and international routes. Nepal has several airlines — Nepal Airlines, Buddha Air, Yeti Airlines, Saurya Airlines, Shree Airlines — and pilots are also needed for charter flights, air ambulance, and cargo operations. Becoming a pilot is expensive and demanding but offers an exciting career with good pay and prestige. Nepal's challenging terrain (mountainous airports like Lukla, Jomsom) makes Nepali pilots among the most skilled in the world. Pilots need excellent vision, quick decision-making, discipline, and extensive training. The career path involves obtaining a Commercial Pilot License (CPL), accumulating flight hours, and progressing through airline ranks from First Officer to Captain.",
		DailyTasks: []string{"Conduct pre-flight checks of aircraft systems", "Review flight plans, weather conditions, and route information", "Communicate with air traffic control for takeoff and landing", "Operate aircraft controls during flight", "Monitor instruments and systems during flight", "Brief cabin crew and ensure passenger safety", "Complete post-flight documentation and reports"},
		Skills: []string{"Aircraft operation and systems knowledge", "Navigation and flight planning", "Communication with ATC (air traffic control)", "Decision-making under pressure", "Crew resource management and teamwork", "Technical understanding of aviation", "Discipline and adherence to procedures", "Good vision and physical fitness"},
		SalaryMin: 600000, SalaryMax: 5000000, Difficulty: 5, FutureProof: 55,
		EducationReq: "+2 (Science stream) required. Commercial Pilot License (CPL) from CAAN-approved flying school. Ground training in Nepal (Civil Aviation Academy) or abroad. Minimum 200 flight hours for CPL. Class 1 medical certificate. ATPL (Airline Transport Pilot License) required for captains. Type rating on specific aircraft (A320, ATR, etc.).",
		Outlook: "Aviation in Nepal is growing with increasing domestic and international flights. Pilot demand follows airline fleet expansion. Foreign employment opportunities exist for experienced pilots. The career is competitive with high initial cost (CPL training costs 4-8 million NPR). Nepali pilots with mountain flying experience are highly regarded internationally. Automation and drone technology may impact the profession long-term.",
		Tags: []string{"transport", "aviation", "pilot", "prestigious"},
		Resources: []resourceSeed{
			{Title: "Civil Aviation Authority of Nepal (CAAN)", URL: "https://www.caan.gov.np", Description: "Aviation regulation, pilot licensing, and career information"},
			{Title: "Nepal Airlines Corporation", URL: "https://www.nepalairlines.com.np", Description: "National flag carrier airline careers"},
			{Title: "Buddha Air Careers", URL: "https://www.buddhaair.com", Description: "Leading domestic airline - pilot career opportunities"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Excel in science subjects and get medical clearance", Description: "Focus on Physics, Mathematics, and English in your +2 science stream. These subjects are fundamental for aviation theory. Pass the Class 1 medical examination conducted by CAAN-approved doctors. Medical requirements include: vision (6/6 uncorrected or correctable to 6/6), no color blindness, good hearing, no chronic conditions, and overall physical fitness. The medical standard is high — only about 30% of applicants pass initially. If you have any medical condition that might disqualify you, consult an aviation medical examiner before investing in training. Good English communication skills are essential — aviation operates in English worldwide.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "CAAN - Pilot License Information", URL: "https://www.caan.gov.np"},
				{Title: "Class 1 medical examination for pilots (YouTube)", URL: "https://www.youtube.com/results?search_query=class+1+medical+examination+for+pilots"},
				{Title: "Aviation English proficiency requirements", URL: "https://www.youtube.com/results?search_query=aviation+english+proficiency+test"},
			}},
			{StepNumber: 2, Title: "Complete commercial pilot training", Description: "Enroll in a CAAN-approved flying school — either in Nepal (Bharatpur, Nepalgunj) or abroad (USA, Philippines, India, South Africa). Training includes: ground school (aviation theory, meteorology, navigation, air law, aircraft systems), flight training (dual and solo flying, circuits, navigation, instrument flying), and obtaining licenses (Private Pilot License first, then Commercial Pilot License). A CPL requires minimum 200 flight hours plus ground school. The total cost is 4-8 million NPR. Many student pilots take loans or get family support. Flying training is demanding — you must pass written exams, flight tests, and medical checks throughout.",
			Duration: "12-18 months", Links: []roadmapLink{
				{Title: "CAAN-approved flying schools in Nepal", URL: "https://www.caan.gov.np"},
				{Title: "Pilot training in the USA for Nepali students (YouTube)", URL: "https://www.youtube.com/results?search_query=nepali+student+pilot+training+usa"},
				{Title: "Commercial Pilot License - what to expect", URL: "https://www.youtube.com/results?search_query=commercial+pilot+license+training+guide"},
			}},
			{StepNumber: 3, Title: "Get flight instructor rating or build hours", Description: "After CPL, most pilots work as flight instructors to build the 1500+ flight hours required by airlines. A Flight Instructor (FI) rating allows you to teach student pilots while accumulating hours. Flight instructors earn modestly but gain valuable experience. Some pilots build hours through charter flying, aerial photography, or cargo operations. Building hours takes 1-3 years depending on opportunities. During this time, continue studying for airline exams, type ratings, and ATPL subjects. Build relationships with airline pilots and industry contacts. The aviation community in Nepal is small — your reputation and connections matter. Persistence is key — many aspiring pilots face delays and setbacks.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Flight instructor career path (YouTube)", URL: "https://www.youtube.com/results?search_query=flight+instructor+career+path+nepal"},
				{Title: "Building flight hours for airline requirements", URL: "https://www.youtube.com/results?search_query=how+to+build+flight+hours+for+airline+pilot"},
				{Title: "Nepali airline pilot job requirements", URL: "https://www.caan.gov.np"},
			}},
			{StepNumber: 4, Title: "Apply for airline First Officer positions", Description: "When you meet the minimum hour requirements (typically 1500+ hours), apply for First Officer positions at Nepali airlines. The application process includes: CV screening, technical exams, simulator assessment, group discussion, and interview. Airlines look for technical competence, crew resource management skills, and attitude. Nepali airlines prefer pilots with mountain flying experience. If selected, you undergo type rating training on the airline's aircraft (ATR, Jetstream, or Airbus for international). Type rating training is intensive — you learn every system of the aircraft. As a First Officer, you assist the Captain and continue learning. The first year is probation.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Nepali airline pilot recruitment process", URL: "https://www.youtube.com/results?search_query=nepali+airline+pilot+recruitment+process"},
				{Title: "Type rating training for ATR/A320 (YouTube)", URL: "https://www.youtube.com/results?search_query=atr+type+rating+training+overview"},
				{Title: "Buddha Air pilot career", URL: "https://www.buddhaair.com"},
			}},
			{StepNumber: 5, Title: "Progress to Captain and upgrade qualifications", Description: "After accumulating 3000-5000 hours as First Officer, take the Captain upgrade exams and simulator checks. Captains have ultimate responsibility for the aircraft and everyone on board. The upgrade requires leadership skills, decision-making ability, and operational experience. Continue training throughout your career — every six months, pilots take simulator checks and medical exams. Maintain your ATPL (Airline Transport Pilot License). Some pilots pursue additional qualifications: instructor ratings, examiner ratings, or management qualifications. International flying requires additional licenses and certifications. Captain is the highest operational rank for pilots.",
			Duration: "5-10 years", Links: []roadmapLink{
				{Title: "Captain upgrade process in airlines (YouTube)", URL: "https://www.youtube.com/results?search_query=captain+upgrade+process+airline+pilot"},
				{Title: "ATPL license requirements and exams", URL: "https://www.youtube.com/results?search_query=atpl+license+requirements+guide"},
				{Title: "Recurrent training and simulator checks for pilots", URL: "https://www.youtube.com/results?search_query=pilot+recurrent+training+simulator"},
			}},
			{StepNumber: 6, Title: "Become a senior captain or move into aviation management", Description: "Senior captains with 10,000+ hours are the most experienced pilots. Some become check pilots (conducting exams for other pilots), training captains (training new pilots and type ratings), or move into aviation management (flight operations manager, chief pilot, safety manager). Others transition to international airlines for higher pay. Some experienced pilots become aviation consultants, accident investigators, or aviation educators. Flying in Nepal, especially mountain flying, makes you one of the most skilled pilots in the world. The responsibility of a pilot is immense — every flight, you carry the lives of passengers and crew in your hands. It is a career of constant learning, discipline, and pride.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Aviation management career path (YouTube)", URL: "https://www.youtube.com/results?search_query=aviation+management+career+nepal"},
				{Title: "International pilot career with Nepali license", URL: "https://www.youtube.com/results?search_query=nepali+pilot+international+airline+career"},
				{Title: "Aviation safety and accident investigation", URL: "https://www.caan.gov.np"},
			}},
		},
	}
}

func logisticsManager() careerSeed {
	return careerSeed{
		CategoryName: "Transport & Logistics", CategorySlug: "transport-logistics", CategoryIcon: "🚚",
		Title: "Logistics Manager", Slug: "logistics-manager",
		Summary: "Logistics managers plan and coordinate the movement of goods — from suppliers to businesses to customers. They manage transportation, warehousing, and supply chain operations.",
		Description: "A logistics manager oversees the flow of goods — from sourcing raw materials to delivering finished products to customers. They manage transportation (trucks, shipping, air cargo), warehousing, inventory, and supply chain coordination. Nepal's logistics sector includes: freight forwarding, warehousing, distribution, courier services, and supply chain management for import/export businesses. With Nepal's landlocked geography, logistics is complex and critical. Logistics managers work for manufacturers, trading companies, freight forwarders, and e-commerce businesses. The field requires organizational skills, problem-solving ability, and knowledge of transportation and customs regulations. E-commerce growth and trade expansion are driving demand for logistics professionals.",
		DailyTasks: []string{"Plan transportation routes and schedules for shipments", "Coordinate with freight forwarders, trucking companies, and customs agents", "Track shipments and resolve delays or issues", "Manage warehouse operations — receiving, storage, dispatch", "Monitor inventory levels and coordinate replenishment", "Prepare shipping documentation (bills of lading, customs forms)", "Analyze logistics costs and find efficiency improvements"},
		Skills: []string{"Transportation management (road, air, sea freight)", "Warehouse operations and inventory management", "Customs clearance and import/export documentation", "Supply chain planning and coordination", "Data analysis and cost optimization", "Vendor and carrier relationship management", "Problem-solving and crisis management", "Computer skills (WMS, ERP systems, Excel)"},
		SalaryMin: 350000, SalaryMax: 1500000, Difficulty: 3, FutureProof: 70,
		EducationReq: "Bachelor's in Business Management, Supply Chain Management, or Logistics. Master's in Logistics preferred for senior roles. Certification from CILT (Chartered Institute of Logistics and Transport) or CSCP (Certified Supply Chain Professional). Experience in transportation or warehouse operations essential.",
		Outlook: "Logistics is growing with Nepal's trade, manufacturing, and e-commerce sectors. The Nepal Logistics Sector Strategy and new transport corridors are creating opportunities. E-commerce logistics (last-mile delivery) is a rapidly growing segment. Experienced logistics managers are in high demand. Supply chain disruptions (COVID, global events) have highlighted the importance of logistics.",
		Tags: []string{"transport", "logistics", "supply chain", "management"},
		Resources: []resourceSeed{
			{Title: "Merojob Logistics Jobs", URL: "https://www.merojob.com", Description: "Find logistics manager jobs in Nepal"},
			{Title: "Nepal Freight Forwarders Association", URL: "https://www.neffa.org.np", Description: "Logistics and freight forwarding industry in Nepal"},
			{Title: "CILT Nepal (Chartered Institute of Logistics & Transport)", URL: "https://www.ciltnepal.org.np", Description: "Logistics professional development and certification in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn logistics and supply chain basics", Description: "Study the fundamentals of logistics: transportation modes (road, air, sea), warehousing, inventory management, supply chain planning, and customs procedures. Nepal's logistics context is unique — landlocked, dependent on India for transit, with challenging geography. Learn about Nepal's trade routes (Tatopani/Rasuwagadhi to China, Birgunj/Bhairahawa to India), customs clearance process, and the role of freight forwarders and customs agents. Understanding how goods move from supplier to customer is the foundation of logistics. Read about supply chain best practices and study real logistics operations. Read logistics industry publications and follow developments in Nepal's trade and transport sector.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Supply chain management basics (YouTube)", URL: "https://www.youtube.com/results?search_query=supply+chain+management+basics+for+beginners"},
				{Title: "Nepal trade and transit routes explained", URL: "https://www.youtube.com/results?search_query=nepal+trade+route+explained"},
				{Title: "NEFFA - Nepal freight forwarding industry", URL: "https://www.neffa.org.np"},
			}},
			{StepNumber: 2, Title: "Get education in logistics and supply chain", Description: "Pursue a degree in Business Management with focus on logistics or Supply Chain Management. Study: transportation economics, warehouse design and operations, inventory management (EOQ, safety stock, ABC analysis), procurement, and global supply chains. Learn about customs regulations, Incoterms, and international trade documentation. Get certified: CILT (Chartered Institute of Logistics and Transport) certification is globally recognized. CSCP and CPIM certifications from APICS are also valuable. Learn ERP systems (SAP, Oracle) and warehouse management systems (WMS). Logistics increasingly relies on technology — TMS (Transportation Management Systems) and WMS are essential tools.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "CILT logistics certification programs", URL: "https://www.ciltnepal.org.np"},
				{Title: "Supply chain management courses (YouTube)", URL: "https://www.youtube.com/results?search_query=supply+chain+management+course+nepal"},
				{Title: "ERP systems in logistics - SAP overview", URL: "https://www.youtube.com/results?search_query=sap+logistics+module+overview"},
			}},
			{StepNumber: 3, Title: "Work in logistics operations", Description: "Start in an entry-level logistics role: logistics coordinator, warehouse assistant, transportation clerk, or documentation officer. Learn the day-to-day operations: processing orders, arranging transportation, handling customs documentation, tracking shipments, and coordinating with warehouses. Learn by doing — logistics is a practical field. Understand the paperwork: bill of lading, packing list, commercial invoice, certificate of origin, customs declaration. Build relationships with transporters, warehouse operators, and customs agents. In Nepal, personal relationships matter in logistics. Learn to handle problems — shipment delays, customs holds, damages. Each problem teaches you how to prevent it in the future.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Find logistics jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Logistics documentation guide - bill of lading (YouTube)", URL: "https://www.youtube.com/results?search_query=bill+of+lading+guide+nepal"},
				{Title: "Customs clearance process in Nepal", URL: "https://www.youtube.com/results?search_query=customs+clearance+nepal+process"},
			}},
			{StepNumber: 4, Title: "Specialize in logistics function or industry", Description: "Specialize in a logistics function: transportation management (managing truck fleets, route planning, carrier contracts), warehouse management (receiving, storage, picking, packing, shipping), inventory planning (demand forecasting, stock optimization), or international logistics (import/export, customs compliance, freight forwarding). Specialize in an industry: pharmaceutical logistics (cold chain management), perishable goods logistics (temperature-controlled transport), e-commerce logistics (last-mile delivery, reverse logistics), or construction logistics (heavy equipment, bulk materials). Industry specialization makes you more valuable. Cold chain logistics for vaccines and perishables is a growing specialty in Nepal.",
			Duration: "2-3 years", Links: []roadmapLink{
				{Title: "Cold chain logistics management (YouTube)", URL: "https://www.youtube.com/results?search_query=cold+chain+logistics+management+nepal"},
				{Title: "E-commerce last mile delivery operations", URL: "https://www.youtube.com/results?search_query=ecommerce+last+mile+delivery+logistics"},
				{Title: "Warehouse management systems (WMS) training", URL: "https://www.youtube.com/results?search_query=warehouse+management+system+basics"},
			}},
			{StepNumber: 5, Title: "Move into logistics management", Description: "With experience, move into logistics manager roles. Logistics managers oversee the entire logistics function for a company — transportation, warehousing, inventory, and team management. Responsibilities include: budget management, vendor selection and negotiation, process improvement, and reporting to senior management. Develop strong analytical skills — logistics managers use data to make decisions (cost analysis, performance metrics, route optimization). Learn to use logistics analytics tools and create dashboards. Build a high-performing team — hire good people, provide training, and create efficient processes. A logistics manager's performance is measured by cost, service level, and reliability.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Logistics KPI metrics and analysis (YouTube)", URL: "https://www.youtube.com/results?search_query=logistics+KPI+metrics+explained"},
				{Title: "Transportation management best practices", URL: "https://www.youtube.com/results?search_query=transportation+management+best+practices"},
				{Title: "Logistics cost optimization strategies", URL: "https://www.youtube.com/results?search_query=logistics+cost+reduction+strategies"},
			}},
			{StepNumber: 6, Title: "Advance to senior supply chain leadership", Description: "Senior logistics professionals become Supply Chain Directors, Logistics Head, or VP Supply Chain. They develop supply chain strategy, manage large teams, and drive transformation. Some become consultants, advising companies on logistics optimization. Others start their own logistics companies — freight forwarding, warehousing, or last-mile delivery businesses. Nepal's logistics sector has significant opportunities as trade grows and supply chains modernize. E-commerce logistics (delivery services for online shopping) is a high-growth segment. Supply chain logistics keeps Nepal's economy moving — every product you use was delivered through a logistics network. Your work as a logistics manager ensures that goods reach people who need them.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Supply chain director - career path (YouTube)", URL: "https://www.youtube.com/results?search_query=supply+chain+director+career+path"},
				{Title: "Starting a logistics company in Nepal", URL: "https://www.merojob.com"},
				{Title: "Logistics technology and digital transformation", URL: "https://www.youtube.com/results?search_query=logistics+digital+transformation+nepal"},
			}},
		},
	}
}
func publicVehicleDriver() careerSeed {
	return careerSeed{
		CategoryName: "Transport & Logistics", CategorySlug: "transport-logistics", CategoryIcon: "🚌",
		Title: "Public Vehicle Driver", Slug: "public-vehicle-driver",
		Summary: "Public vehicle drivers operate buses, minibuses, tempos, and other vehicles that transport passengers on fixed or flexible routes within cities and across Nepal.",
		Description: "A public vehicle driver transports passengers in buses, minibuses, tempos, microbuses, and other public transport vehicles. Nepal's public transport system is extensive — from city routes (Kathmandu valley, major cities) to intercity and rural routes that connect even remote areas. Drivers need a valid driving license for the appropriate vehicle category, knowledge of traffic rules, and good driving skills. Public transport driving is a demanding job — long hours, traffic congestion, road hazards (especially on mountain roads), and dealing with diverse passengers. It provides stable income for those who own their vehicles or work for transport companies. In Nepal, many drivers own their vehicles and operate on cooperative or association-based routes. The job requires patience, alertness, and excellent driving skills.",
		DailyTasks: []string{"Drive assigned route according to schedule", "Collect fares from passengers and issue tickets", "Check vehicle condition before and after trips", "Follow traffic rules and drive safely", "Assist passengers with luggage and boarding", "Handle breakdowns and emergencies on the road", "Clean and maintain the vehicle"},
		Skills: []string{"Safe driving skills for Nepali road conditions", "Knowledge of traffic rules and regulations", "Vehicle maintenance knowledge (basic repairs)", "Customer service and dealing with passengers", "Route knowledge of cities or regions", "Patience and stress management", "Physical stamina for long driving hours", "Basic math for fare collection"},
		SalaryMin: 180000, SalaryMax: 600000, Difficulty: 2, FutureProof: 45,
		EducationReq: "SEE pass. Valid driving license for appropriate vehicle category (Bus/Heavy Vehicle). Transport operator license from Department of Transport Management. Route permit and vehicle registration. Defensive driving training. Basic vehicle mechanics knowledge helpful.",
		Outlook: "Public transport is essential in Nepal and will continue to need drivers. Route associations and cooperatives control many routes. Vehicle ownership (owning your bus/microbus) significantly increases income. Electric vehicles (e-rickshaws, e-microbuses) are growing — drivers who can operate and maintain EVs have an advantage. Long-distance routes connect Kathmandu to all districts — experienced drivers for these routes are valued.",
		Tags: []string{"transport", "driving", "public service", "hands-on"},
		Resources: []resourceSeed{
			{Title: "Department of Transport Management Nepal", URL: "https://www.dotm.gov.np", Description: "Driving licenses, route permits, and transport regulations"},
			{Title: "Merojob Transport Jobs", URL: "https://www.merojob.com", Description: "Find driver and transport jobs in Nepal"},
			{Title: "Defensive driving tips (YouTube)", URL: "https://www.youtube.com/results?search_query=defensive+driving+tips+nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn to drive and obtain licenses", Description: "Learn driving from a licensed driving school. Get a learner's permit, practice driving, and obtain a driving license for two-wheelers first (if you do not have one), then light vehicles, and eventually heavy vehicles (bus/truck). Department of Transport Management (DOTM) conducts driving tests. For public vehicle driving, you need a professional license for the appropriate vehicle category. Learn traffic rules, road signs, and defensive driving techniques. Nepali roads are challenging — narrow, congested in cities, and dangerous on mountain routes. Safe driving skills are essential for public transport. Good drivers anticipate hazards, drive smoothly, and protect passengers. Driving experience on different road types is important.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "DOTM - Driving License Information", URL: "https://www.dotm.gov.np"},
				{Title: "Driving test tips for Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=driving+test+nepal+tips+and+tricks"},
				{Title: "Defensive driving course for professionals", URL: "https://www.youtube.com/results?search_query=defensive+driving+course+nepal"},
			}},
			{StepNumber: 2, Title: "Get professional driving experience", Description: "Start driving smaller vehicles (tempo, minibus) before moving to larger buses. Work as a driver for an established transport company or route association. Learn the routes — Nepal's roads have specific challenges: narrow streets in cities, blind curves on highways, landslide-prone areas during monsoon, and high altitude passes. Learn vehicle maintenance — check oil, water, tire pressure, brakes daily. Know what to do in breakdowns, accidents, and emergencies. Develop good customer service — passengers appreciate drivers who are helpful, safe, and courteous. Build a reputation as a reliable and safe driver. In the transport business, reputation determines your income.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Bus driving skills for mountain roads (YouTube)", URL: "https://www.youtube.com/results?search_query=mountain+road+bus+driving+tips+nepal"},
				{Title: "Basic vehicle maintenance for drivers", URL: "https://www.youtube.com/results?search_query=basic+vehicle+maintenance+for+bus+drivers"},
				{Title: "Dealing with passengers - tips for public transport drivers", URL: "https://www.youtube.com/results?search_query=customer+service+for+public+transport+drivers"},
			}},
			{StepNumber: 3, Title: "Understand the transport route and business system", Description: "Nepal's public transport system operates through route associations, cooperatives, and individual owners. Learn how your route works — who manages the route, how fares are set, how schedules are managed, and how income is shared. Understand the costs: vehicle maintenance, fuel, insurance, route permit fees, and association fees. If you drive for someone else (vehicle owner), understand your salary or commission arrangement. Some drivers work on a daily profit-sharing basis — keeping a percentage of daily collections. Building relationships with route association officials, fellow drivers, and mechanics is essential. The public transport business is relationship-driven.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Nepal transport route associations explained", URL: "https://www.youtube.com/results?search_query=nepal+public+transport+route+association"},
				{Title: "Transport business economics in Nepal", URL: "https://www.merojob.com"},
				{Title: "Vehicle ownership and financing for transport", URL: "https://www.youtube.com/results?search_query=financing+public+transport+vehicle+nepal"},
			}},
			{StepNumber: 4, Title: "Specialize in route or vehicle type", Description: "Specialize in: city routes (microbuses, tempos in Kathmandu, Pokhara, etc.), intercity routes (long-distance buses connecting major cities), rural routes (buses connecting district headquarters to villages — requires mountain driving skills), tourist transport (tourist buses, airport transfers — higher service standards and English helpful), or school/college transport (fixed daily routes with students). Long-distance bus driving pays better but requires long hours away from home. Tourist transport, especially to trekking trailheads, requires knowledge of tourism and good service skills. Specialization in tourist transport can lead to higher earnings and tips.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Tourist bus driving in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=tourist+bus+driving+nepal+guide"},
				{Title: "Mountain bus routes and driving challenges", URL: "https://www.youtube.com/results?search_query=nepal+mountain+bus+route+driving"},
				{Title: "School transport management and safety", URL: "https://www.youtube.com/results?search_query=school+bus+safety+nepal"},
			}},
			{StepNumber: 5, Title: "Buy your own vehicle or join a cooperative", Description: "Many successful drivers purchase their own vehicle and operate under a route association or cooperative. Vehicle ownership significantly increases income — you earn both driver income and vehicle owner income. A used bus or microbus can cost 1-5 million NPR depending on age and condition. Financing is available through banks and cooperatives. Joining a transport cooperative gives you access to route permits, scheduled operations, and collective bargaining. As an owner-driver, you manage: vehicle maintenance, driver relief (if you hire someone for alternate shifts), accounting, and regulatory compliance. Owner-drivers who maintain their vehicles well and operate reliably earn the most.",
			Duration: "2-5 years", Links: []roadmapLink{
				{Title: "Buying a bus for public transport Nepal", URL: "https://www.youtube.com/results?search_query=buy+used+bus+nepal+public+transport"},
				{Title: "Transport cooperatives in Nepal", URL: "https://www.youtube.com/results?search_query=transport+cooperative+nepal+benefits"},
				{Title: "Vehicle financing options for drivers", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 6, Title: "Expand your transport business", Description: "Owner-drivers can expand: purchase additional vehicles and hire drivers, start a transport company with multiple routes, diversify into goods transport (trucking), or expand into related businesses (auto workshop, spare parts, tire shop). Experienced drivers also become trainers — teaching new drivers safe driving techniques, especially for mountain routes. Some become route association leaders, managing route schedules and representing drivers to government authorities. Public transport is the backbone of mobility in Nepal. Your driving connects people to their jobs, families, education, and healthcare. As an experienced driver or transport business owner, you provide an essential service to your community.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Growing a transport business in Nepal", URL: "https://www.youtube.com/results?search_query=grow+transport+business+nepal"},
				{Title: "Trucking and goods transport business", URL: "https://www.merojob.com"},
				{Title: "Route association management and leadership", URL: "https://www.youtube.com/results?search_query=transport+route+association+management+nepal"},
			}},
		},
	}
}

func deliveryService() careerSeed {
	return careerSeed{
		CategoryName: "Transport & Logistics", CategorySlug: "transport-logistics", CategoryIcon: "📦",
		Title: "Delivery Service (Courier/Rider)", Slug: "delivery-service",
		Summary: "Delivery riders and couriers deliver packages, food, documents, and goods to customers — using motorcycles, scooters, or bicycles in cities across Nepal.",
		Description: "A delivery service provider (delivery rider, courier) delivers items to customers — food from restaurants, packages from online shopping, documents for businesses, and goods from stores. This sector has exploded in Nepal with the growth of food delivery apps (Foodmandu, Bhoj, Delivery Nepal), e-commerce (Daraz, Sastodeal), and courier services (DHL, FedEx, local courier companies). Delivery riders typically use motorcycles or scooters for fast urban delivery. The job requires a valid driving license, navigation skills, and physical fitness. Delivery riders can work for delivery platforms, courier companies, or start their own independent delivery service. Flexible hours make it attractive for students or those wanting part-time work. Good delivery riders earn through base pay, delivery fees, and tips.",
		DailyTasks: []string{"Receive delivery orders through app or dispatch", "Pick up items from restaurants, stores, or warehouses", "Plan efficient delivery routes", "Deliver items to customers promptly and safely", "Collect payments (if cash on delivery)", "Maintain delivery vehicle (motorcycle/scooter)", "Handle customer communication regarding delivery status"},
		Skills: []string{"Motorcycle or scooter riding with valid license", "City navigation and route planning", "Time management and punctuality", "Customer service and communication", "Basic record keeping (delivery logs, payments collected)", "Physical fitness for carrying items", "Smartphone use for delivery apps and maps", "Problem-solving for delivery issues"},
		SalaryMin: 150000, SalaryMax: 600000, Difficulty: 1, FutureProof: 65,
		EducationReq: "SEE pass. Valid motorcycle driving license. Smartphone with internet. Delivery experience helpful but not required. Food handler's permit may be needed for food delivery. Basic English for app navigation.",
		Outlook: "Delivery services are growing rapidly with food delivery apps and e-commerce. The sector offers flexible earning opportunities. App-based delivery (Foodmandu, Bhoj, Daraz) provides steady work. Courier companies (DHL, FedEx, local) employ delivery personnel. Electric scooters are increasingly used for deliveries — lower operating costs. The gig economy model means income depends on hours worked and orders delivered.",
		Tags: []string{"transport", "delivery", "gig economy", "flexible"},
		Resources: []resourceSeed{
			{Title: "Foodmandu - Delivery Partner", URL: "https://www.foodmandu.com", Description: "Nepal's leading food delivery platform - delivery rider opportunities"},
			{Title: "Bhoj Delivery - Rider Application", URL: "https://www.bhoj.com.np", Description: "Food delivery service - rider career information"},
			{Title: "Daraz Logistics - Delivery Jobs", URL: "https://www.daraz.com.np", Description: "E-commerce delivery and logistics careers in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Get a motorcycle and license", Description: "Obtain a valid motorcycle driving license from DOTM. Purchase or arrange access to a reliable motorcycle or scooter. The vehicle should be fuel-efficient and low-maintenance — popular models for delivery in Nepal include 100-150cc motorcycles (Hero Splendor, Honda Shine, Bajaj Platina). Learn city navigation — know the streets, shortcuts, and traffic patterns of your city. Learn to use maps apps (Google Maps) and delivery apps. Customer service skills matter — being polite and professional when interacting with customers. Delivery is a simple entry-level job, but good riders who are reliable, fast, and professional earn more and get better assignments.",
			Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Motorbike license in Nepal - process (YouTube)", URL: "https://www.youtube.com/results?search_query=motorbike+license+nepal+process+2024"},
				{Title: "Best motorcycle for delivery business Nepal", URL: "https://www.youtube.com/results?search_query=best+bike+for+delivery+nepal"},
				{Title: "City navigation tips for delivery riders", URL: "https://www.youtube.com/results?search_query=delivery+rider+navigation+tips"},
			}},
			{StepNumber: 2, Title: "Start as a delivery rider for a platform", Description: "Register with a delivery platform as a delivery partner. Foodmandu, Bhoj, and other apps have onboarding processes. Submit your documents (license, vehicle registration, citizenship). Complete any training provided by the platform. Start accepting orders. Learn the system — which orders to accept, how to handle delays, how to communicate with customers. Build your efficiency — plan routes to minimize travel time, know which restaurants take longer, learn apartment building layouts. Riders who deliver quickly and maintain high ratings get priority for orders. Delivery is a numbers game — the more deliveries you complete, the more you earn. Aim for consistent daily delivery targets.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Foodmandu rider registration guide (YouTube)", URL: "https://www.youtube.com/results?search_query=foodmandu+rider+registration+nepal"},
				{Title: "Tips to maximize delivery earnings", URL: "https://www.youtube.com/results?search_query=how+to+maximize+food+delivery+earnings"},
				{Title: "Delivery app ratings - how to maintain high scores", URL: "https://www.youtube.com/results?search_query=delivery+app+rider+rating+tips"},
			}},
			{StepNumber: 3, Title: "Work with multiple platforms or courier companies", Description: "To maximize income, register with multiple delivery platforms. Work for food delivery apps during peak meal times and offer courier services during slower hours. Apply to courier companies for package delivery work. Some riders combine: morning document delivery for offices, lunch food delivery, afternoon package runs, and dinner food delivery. Learn about different delivery types: food (requires care with packaging and temperature), documents (requires careful handling), and packages (requires space and possibly different vehicle). Build relationships with regular customers and businesses — they may request you directly for deliveries. Regular clients provide stable income.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Multiple delivery apps - balancing strategy (YouTube)", URL: "https://www.youtube.com/results?search_query=work+multiple+delivery+apps+strategy"},
				{Title: "Courier delivery jobs in Nepal", URL: "https://www.merojob.com"},
				{Title: "Document and package delivery business tips", URL: "https://www.youtube.com/results?search_query=document+delivery+service+nepal"},
			}},
			{StepNumber: 4, Title: "Build your own delivery client base", Description: "Beyond app-based work, build your own delivery client base. Distribute your business card to restaurants, shops, and offices. Offer reliable same-day delivery services. Many businesses need occasional deliveries but do not use formal courier services. Offer a package deal — weekly or monthly delivery contracts for regular clients. Build a reputation for reliability — show up on time, handle packages carefully, communicate clearly. Word of mouth brings more clients. Develop a simple system: phone number for orders, a log of deliveries, and transparent pricing. Independent delivery service offers more control over earnings and hours.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Starting a local delivery business (YouTube)", URL: "https://www.youtube.com/results?search_query=start+local+delivery+business+nepal"},
				{Title: "Marketing your delivery service locally", URL: "https://www.youtube.com/results?search_query=marketing+delivery+service+nepal"},
				{Title: "Pricing for independent delivery services", URL: "https://www.youtube.com/results?search_query=delivery+service+pricing+guide"},
			}},
			{StepNumber: 5, Title: "Expand into a small delivery fleet", Description: "As your client base grows, hire other riders to work for you. You become the dispatcher — receiving orders and assigning them to your riders. Manage a small fleet of 2-5 riders serving regular business clients. Manage: rider scheduling, delivery tracking, client relationships, and accounting. Fleet management requires organizational skills and reliability. You earn a margin on each delivery. Start with one additional rider and expand based on demand. Focus on business-to-business deliveries (regular, predictable income) rather than relying solely on on-demand food delivery.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Building a delivery fleet business (YouTube)", URL: "https://www.youtube.com/results?search_query=building+a+delivery+fleet+business"},
				{Title: "Managing delivery riders - tips", URL: "https://www.youtube.com/results?search_query=manage+delivery+rider+team"},
				{Title: "Delivery business software and tools", URL: "https://www.youtube.com/results?search_query=delivery+business+management+app"},
			}},
			{StepNumber: 6, Title: "Start a courier company or logistics service", Description: "Experienced delivery operators can start a registered courier company serving local or national routes. Obtain necessary permits from DOTM and register your company. Offer services: same-day delivery, overnight delivery, specialized delivery (pharmaceuticals, perishables), or e-commerce fulfillment. Build a brand — reliable delivery service with tracking and customer support. Courier businesses require investment in vehicles, technology, staff, and marketing. Some successful riders transition to logistics technology — developing apps or platforms for delivery services. The delivery and logistics sector in Nepal is growing — early entrants with good service build valuable businesses.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Starting a courier company in Nepal", URL: "https://www.merojob.com"},
				{Title: "E-commerce fulfillment and logistics business", URL: "https://www.youtube.com/results?search_query=ecommerce+fulfillment+business+nepal"},
				{Title: "Delivery business regulations and permits Nepal", URL: "https://www.dotm.gov.np"},
			}},
		},
	}
}

func warehouseManager() careerSeed {
	return careerSeed{
		CategoryName: "Transport & Logistics", CategorySlug: "transport-logistics", CategoryIcon: "🏭",
		Title: "Warehouse Manager", Slug: "warehouse-manager",
		Summary: "Warehouse managers oversee storage and distribution operations — receiving goods, managing inventory, coordinating shipments, and leading warehouse teams.",
		Description: "A warehouse manager is responsible for all activities in a warehouse: receiving goods from suppliers, storing products properly, managing inventory accuracy, picking and packing orders, and shipping goods to customers. Warehouse managers work for manufacturers, distributors, retailers, logistics companies, and e-commerce businesses. Nepal's growing trade and e-commerce sectors are increasing demand for organized warehousing. Good warehouse management ensures that products are stored safely, inventory is accurate, and orders are shipped on time. The role requires organizational skills, team leadership, knowledge of warehouse systems, and attention to safety. Modern warehouses use technology (WMS, barcode scanners, forklifts) but Nepal still has many traditional warehouses. Warehouse managers need to work with both modern and traditional systems.",
		DailyTasks: []string{"Oversee receiving of incoming goods and verify quantities", "Organize warehouse layout for efficient storage and picking", "Supervise warehouse staff — assign tasks, monitor performance", "Maintain inventory accuracy through regular cycle counts", "Coordinate with logistics team for inbound and outbound shipments", "Ensure safety standards and proper equipment maintenance", "Prepare warehouse reports — inventory levels, productivity, issues"},
		Skills: []string{"Warehouse operations and layout planning", "Inventory management and stock control", "Team supervision and training", "Safety management and regulatory compliance", "Computer skills (WMS, Excel, inventory software)", "Forklift and material handling equipment operation", "Problem-solving and process improvement", "Physical fitness for warehouse environment"},
		SalaryMin: 300000, SalaryMax: 1000000, Difficulty: 3, FutureProof: 60,
		EducationReq: "Bachelor's in Business Management, Logistics, or Supply Chain Management. Warehouse management certification (CILT, APICS). Forklift operation certification. Experience in warehouse operations essential — most managers start as warehouse associates. Safety training (OSHA standards or equivalent).",
		Outlook: "Modern warehousing is growing with e-commerce, retail chains, and manufacturing. Organized warehouses are replacing traditional open storage. Cold storage (for pharmaceuticals, food) is a growing segment. Experienced warehouse managers with WMS knowledge are in demand. Automation (conveyors, sorters) is increasing but human management remains essential. Nepal's logistics modernization creates opportunities.",
		Tags: []string{"transport", "warehouse", "logistics", "management"},
		Resources: []resourceSeed{
			{Title: "Merojob Warehouse Jobs", URL: "https://www.merojob.com", Description: "Find warehouse manager jobs in Nepal"},
			{Title: "CILT Nepal - Warehouse Certification", URL: "https://www.ciltnepal.org.np", Description: "Professional certification for warehouse and logistics management"},
			{Title: "Warehouse management systems guide (YouTube)", URL: "https://www.youtube.com/results?search_query=warehouse+management+system+basics+tutorial"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Start working in warehouse operations", Description: "Begin as a warehouse associate or helper. Learn the basics: receiving goods (checking quantities, quality inspection), storing products (proper placement, FIFO/FEFO rotation), picking orders (finding products, accuracy), packing shipments (proper packaging, labeling), and loading trucks. Learn to operate warehouse equipment — pallet jack, hand truck, forklift (if available). Understand inventory documentation — goods received notes, stock cards, bin location systems. Pay attention to safety — proper lifting techniques, clean aisles, fire safety. Good warehouse associates are organized, careful, and efficient. Build a reputation for accuracy and reliability. Every warehouse manager started on the warehouse floor.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Warehouse operations basics (YouTube)", URL: "https://www.youtube.com/results?search_query=warehouse+operations+basics+for+beginners"},
				{Title: "Forklift operation training basics", URL: "https://www.youtube.com/results?search_query=forklift+training+for+beginners+nepal"},
				{Title: "Inventory management fundamentals - FIFO FEFO", URL: "https://www.youtube.com/results?search_query=fifo+fefo+inventory+management+explained"},
			}},
			{StepNumber: 2, Title: "Learn warehouse management systems and inventory control", Description: "Learn to use Warehouse Management Systems (WMS) and inventory software. Microsoft Excel is essential for inventory tracking and reporting. Learn about inventory management principles: cycle counting (regular partial inventory counts), ABC analysis (classifying items by value), safety stock levels, and reorder points. Study warehouse layout design — how to arrange products for efficient storage and picking. Understand key performance indicators (KPIs): order accuracy, picking speed, inventory accuracy, space utilization. Take courses in supply chain management or logistics. Formal education combined with practical experience positions you for management roles.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "WMS software training - overview (YouTube)", URL: "https://www.youtube.com/results?search_query=warehouse+management+system+training"},
				{Title: "Excel for inventory management", URL: "https://www.youtube.com/results?search_query=excel+for+inventory+management+tutorial"},
				{Title: "Warehouse KPIs and performance measurement", URL: "https://www.youtube.com/results?search_query=warehouse+KPI+metrics+explained"},
			}},
			{StepNumber: 3, Title: "Become a warehouse supervisor or team lead", Description: "Progress to team lead or supervisor role. Supervisor responsibilities include: assigning daily tasks to warehouse associates, monitoring productivity and accuracy, training new staff, ensuring safety compliance, and reporting to the warehouse manager. Learn to manage people — motivate your team, handle performance issues, and create a positive work environment. Good supervisors lead by example — they work alongside their team when needed. Learn to communicate effectively with other departments — procurement (inbound shipments), sales/customer service (outbound orders), and transport (loading schedules). Supervisory experience is the stepping stone to warehouse management.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Warehouse supervisor roles and responsibilities (YouTube)", URL: "https://www.youtube.com/results?search_query=warehouse+supervisor+duties+guide"},
				{Title: "Team leadership in warehouse environment", URL: "https://www.youtube.com/results?search_query=warehouse+team+leadership+tips"},
				{Title: "Safety management in warehouses", URL: "https://www.youtube.com/results?search_query=warehouse+safety+management+standards"},
			}},
			{StepNumber: 4, Title: "Take on full warehouse manager responsibilities", Description: "As warehouse manager, you oversee all operations: receiving, storage, picking, packing, shipping, and returns. Manage warehouse budgets, staffing levels, and equipment maintenance. Ensure inventory accuracy through systematic cycle counting and annual physical inventory. Analyze warehouse processes and implement improvements. Reduce costs while maintaining service levels. Manage warehouse safety program — training, inspections, accident reporting. Build relationships with transport companies and coordinate loading schedules. Report to senior management on warehouse performance. A good warehouse manager runs an efficient, safe, and accurate operation. Your performance is measured by cost per order, accuracy rate, and on-time shipment.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Warehouse manager responsibilities (YouTube)", URL: "https://www.youtube.com/results?search_query=warehouse+manager+daily+duties"},
				{Title: "Warehouse cost reduction strategies", URL: "https://www.youtube.com/results?search_query=warehouse+cost+reduction+ideas"},
				{Title: "Warehouse safety program development", URL: "https://www.youtube.com/results?search_query=warehouse+safety+program+guide"},
			}},
			{StepNumber: 5, Title: "Specialize in warehouse type or technology", Description: "Specialize in: cold storage warehouse (temperature-controlled for food, pharmaceuticals), bonded warehouse (for imported goods awaiting customs clearance), automated warehouse (conveyor systems, automated storage and retrieval systems), hazardous materials warehouse (special safety requirements), or e-commerce fulfillment warehouse (high-volume, many small orders). Each specialization requires specific knowledge. Cold storage requires understanding of temperature control systems and perishable goods handling. Automated warehouses require understanding of material handling systems and maintenance. Specialization in a growing niche (cold chain, e-commerce fulfillment) positions you for higher pay and more opportunities.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Cold storage warehouse management (YouTube)", URL: "https://www.youtube.com/results?search_query=cold+storage+warehouse+management+guide"},
				{Title: "E-commerce fulfillment warehouse operations", URL: "https://www.youtube.com/results?search_query=ecommerce+fulfillment+warehouse+process"},
				{Title: "Bonded warehouse and customs procedures Nepal", URL: "https://www.youtube.com/results?search_query=bonded+warehouse+nepal+procedures"},
			}},
			{StepNumber: 6, Title: "Advance to supply chain or operations leadership", Description: "Senior warehouse professionals become Supply Chain Managers, Operations Managers, or Logistics Directors. They oversee multiple warehouses, design supply chain networks, and develop logistics strategy. Some become consultants, helping companies improve their warehouse operations. Others start their own warehousing business — building and renting warehouse space with management services. Nepal's warehousing sector is modernizing. Organized, professionally managed warehouses with good systems are in demand. Warehouse managers who combine operational expertise with technology and people management skills have strong career prospects. Your work ensures that products are available when and where people need them.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "From warehouse manager to supply chain director (YouTube)", URL: "https://www.youtube.com/results?search_query=supply+chain+director+career+path+warehouse"},
				{Title: "Starting a warehousing business in Nepal", URL: "https://www.merojob.com"},
				{Title: "Warehouse automation and technology trends", URL: "https://www.youtube.com/results?search_query=warehouse+automation+trends+nepal"},
			}},
		},
	}
}
func astrologer() careerSeed {
	return careerSeed{
		CategoryName: "Other Essential Nepal Careers", CategorySlug: "other-essential", CategoryIcon: "🔮",
		Title: "Astrologer (Jyotish)", Slug: "astrologer",
		Summary: "Astrologers study planetary positions and their influence on human affairs. They provide horoscopes, predictions, and guidance for important life decisions.",
		Description: "An astrologer (Jyotish in Nepali) studies the positions and movements of celestial bodies to interpret their influence on human affairs and natural events. Astrology is deeply embedded in Nepali culture and Hindu tradition. Astrologers are consulted for: naming ceremonies (namakaran), marriage matching (kundali milan), starting new businesses (muhurta), house construction (vastu), career guidance, and predicting future trends. Astrologers use Vedic astrology (Jyotish Shastra) based on the sidereal zodiac. Knowledge of Sanskrit, astronomy, and traditional texts (Brihat Parashara Hora Shastra) is valued. Many astrologers in Nepal learn through family tradition (generations of astrologers). Others study formally through Sanskrit universities or gurukul education. Astrology combines religious knowledge, mathematics (calculation of planetary positions), and intuitive interpretation. It remains a respected profession in Nepal.",
		DailyTasks: []string{"Create birth charts (kundali) based on date, time, and place of birth", "Interpret planetary positions and their influence", "Match horoscopes for marriage compatibility", "Determine auspicious dates and times (muhurta) for events", "Advise clients on remedies for planetary afflictions", "Study astrological texts and update calculations", "Maintain client records and astrological charts"},
		Skills: []string{"Knowledge of Vedic astrology principles and texts", "Calculation of planetary positions (panchanga)", "Birth chart (kundali) creation and interpretation", "Understanding of Hindu calendar system (Vikram Samvat)", "Marriage matching (kundali milan) expertise", "Muhurta (auspicious timing) determination", "Consultation and advisory communication", "Ethical practice and client confidentiality"},
		SalaryMin: 120000, SalaryMax: 600000, Difficulty: 3, FutureProof: 30,
		EducationReq: "Traditional training under a guru (gurukul system). Formal study at Sanskrit universities (TU, Nepal Sanskrit University). Knowledge of Jyotish Shastra. Study of Panchanga and ephemeris. Computer skills for astrology software (optional but helpful). Many astrologers learn through family tradition.",
		Outlook: "Astrology remains culturally significant in Nepal for weddings, business starts, and naming ceremonies. Demand is steady but not growing significantly. Young people are less invested in astrology but traditional practices continue. The profession faces competition from modern counseling and online astrology services. Professional astrologers with good reputation maintain steady clienteles.",
		Tags: []string{"other", "astrology", "traditional", "spiritual"},
		Resources: []resourceSeed{
			{Title: "Nepal Sanskrit University - Jyotish Program", URL: "https://www.nsu.edu.np", Description: "Sanskrit education and astrology studies in Nepal"},
			{Title: "Nepal Panchanga Nirnayak Samiti", URL: "https://www.panchanga.org.np", Description: "Official Nepali calendar and astrological calculations"},
			{Title: "Vedic astrology learning resources (YouTube)", URL: "https://www.youtube.com/results?search_query=vedic+astrology+learning+for+beginners+nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn the fundamentals of Vedic astrology", Description: "Begin studying the basics of Vedic astrology: the 12 zodiac signs (rashis), 9 planets (grahas), 27 lunar mansions (nakshatras), and 12 houses (bhavas). Understand the Hindu calendar system (Vikram Samvat, lunar months). Learn to read the panchanga (Nepali calendar/almanac) which contains daily astrological information. Study from books, online resources, or find a teacher. Many Nepali astrologers learned from their fathers or grandfathers who practiced the profession. Sanskrit knowledge is helpful for reading original texts. Start with basic chart interpretation and gradually build deeper knowledge. Astrology is a complex subject that takes years to master.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Vedic astrology basics for beginners (YouTube)", URL: "https://www.youtube.com/results?search_query=vedic+astrology+basics+for+beginners"},
				{Title: "Nepali Panchanga - how to read it", URL: "https://www.panchanga.org.np"},
				{Title: "Sanskrit for astrology - basic terms", URL: "https://www.youtube.com/results?search_query=sanskrit+for+astrology+students"},
			}},
			{StepNumber: 2, Title: "Get formal training or mentorship", Description: "Study under an experienced astrologer (guru) who can teach you the practical aspects. In Nepal, the guru-shishya tradition is the most respected path for learning astrology. Alternatively, enroll in a formal Jyotish program at Nepal Sanskrit University or other institutions. Learn to create accurate birth charts manually — calculation of planetary positions using ephemeris, formulas for ascendant (lagna), and house division. Then learn astrology software that automates calculations. Study predictive techniques: dashas (planetary periods), transits (gochara), and yogas (planetary combinations). Learn marriage matching methodology — the 8 aspects (koota/guna) of compatibility. Deep knowledge takes years of dedicated study.",
			Duration: "3-5 years", Links: []roadmapLink{
				{Title: "Nepal Sanskrit University - Jyotish Faculty", URL: "https://www.nsu.edu.np"},
				{Title: "Manual chart calculation tutorial (YouTube)", URL: "https://www.youtube.com/results?search_query=manual+birth+chart+calculation+vedic+astrology"},
				{Title: "Kundali matching - 8 kootas explained", URL: "https://www.youtube.com/results?search_query=kundali+matching+8+gunas+explained+nepal"},
			}},
			{StepNumber: 3, Title: "Start practicing with real consultations", Description: "Begin offering astrology services to family, friends, and community members — initially free or for minimal fees. Practice makes perfect — the more charts you read, the better your interpretations become. Build confidence in your predictions and advice. Develop your consultation style — how to communicate complex astrological concepts in simple, helpful terms. Keep records of your predictions and follow up to verify their accuracy. Honest self-assessment of your predictions helps you improve. Many experienced astrologers say it takes 10+ years to become truly proficient. Build a reputation for accuracy and ethical practice. Good astrologers help people make better decisions.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Building an astrology practice (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+build+astrology+practice"},
				{Title: "Astrology consultation tips for beginners", URL: "https://www.youtube.com/results?search_query=astrology+consultation+tips"},
				{Title: "Ethical guidelines for astrologers", URL: "https://www.youtube.com/results?search_query=astrology+ethics+and+responsibility"},
			}},
			{StepNumber: 4, Title: "Deepen your knowledge and specialize", Description: "Specialize in a branch of astrology: horary astrology (answering specific questions), medical astrology (health predictions), Muhurta (electional astrology for choosing auspicious times), or Vastu Shastra (the Hindu system of architecture and spatial arrangement). Study advanced predictive techniques: Varshaphal (annual horoscopy), Jaimini astrology, or Tajika system. Many astrologers also learn gemology — recommending gemstones as remedies for planetary afflictions (a significant income source). Specialization in Vastu Shastra is particularly valuable in Nepal for house construction and business premises. Continue studying original texts — Brihat Parashara Hora Shastra, Phaladeepika, and other classics.",
			Duration: "2-5 years", Links: []roadmapLink{
				{Title: "Muhurta - electional astrology guide (YouTube)", URL: "https://www.youtube.com/results?search_query=muhurta+astrology+basics+nepal"},
				{Title: "Vastu Shastra basics for homes and businesses", URL: "https://www.youtube.com/results?search_query=vastu+shastra+basics+nepal"},
				{Title: "Gemstone recommendations in Vedic astrology", URL: "https://www.youtube.com/results?search_query=gemstones+in+vedic+astrology+guide"},
			}},
			{StepNumber: 5, Title: "Build a professional astrology practice", Description: "Set up a dedicated consultation space — either a small office or consultations from home. Establish your reputation through word of mouth. Many astrologers also have an online presence — offering services through social media, websites, or video consultations. Nepali communities abroad (diaspora) are a market for online astrology services. Develop your fee structure — from 500-5000 NPR per consultation depending on your reputation and the service. Offer package services: full birth chart analysis, annual prediction, marriage matching, and specific question answering. Maintain client confidentiality and professional ethics. A respected astrologer can build a stable practice serving generations of families.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Setting up an astrology consultation practice", URL: "https://www.youtube.com/results?search_query=astrology+consultation+setup+nepal"},
				{Title: "Online astrology services - reaching global clients", URL: "https://www.youtube.com/results?search_query=online+astrology+services+nepal"},
				{Title: "Astrology business pricing guide", URL: "https://www.youtube.com/results?search_query=astrology+fee+structure+nepal"},
			}},
			{StepNumber: 6, Title: "Become a respected astrologer and community guide", Description: "Respected astrologers in Nepal are community leaders who are consulted for major life decisions. They may be invited to temples, community events, and even government functions for auspicious timing. Some astrologers teach the next generation, write books or articles on astrology, or appear on television and radio programs discussing astrological predictions. Others combine astrology with other services — vastu consultation, gemstone sales, and religious ceremony guidance. Astrology remains a living tradition in Nepal. Your knowledge of this ancient science can provide guidance and comfort to people navigating life's important decisions.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Teaching astrology - guru to next generation", URL: "https://www.youtube.com/results?search_query=teaching+vedic+astrology+guide"},
				{Title: "Astrology in Nepali media - TV and radio", URL: "https://www.youtube.com/results?search_query=astrology+nepal+tv+program"},
				{Title: "Astrology and spirituality - lifelong learning path", URL: "https://www.youtube.com/results?search_query=advanced+vedic+astrology+learning"},
			}},
		},
	}
}

func handicraft() careerSeed {
	return careerSeed{
		CategoryName: "Other Essential Nepal Careers", CategorySlug: "other-essential", CategoryIcon: "🪆",
		Title: "Handicraft Artisan", Slug: "handicraft-artisan",
		Summary: "Handicraft artisans create traditional Nepali crafts — statues, paintings, textiles, pottery, and decorative items — for local and export markets.",
		Description: "A handicraft artisan creates traditional Nepali crafts using skills passed down through generations. Nepal is famous for: metal statues (brass, copper, bronze Buddha and Hindu deities), Thangka paintings (traditional Buddhist scroll paintings), wood carving (windows, frames, decorative pieces), pottery (clay pots, decorative items), felt products (slippers, bags, decorations), pashmina products (shawls, scarves), and paper products (lokta paper, handmade journals, cards). Handicrafts are a significant export industry for Nepal. Artisans work in their homes, small workshops, or larger manufacturing facilities. The handicraft sector provides employment to thousands of Nepali families, especially in the Kathmandu Valley (Patan for metal crafts, Bhaktapur for pottery, Thimi for masks). The career requires patience, precision, creativity, and dedication to preserving traditional techniques.",
		DailyTasks: []string{"Create handicraft items using traditional techniques", "Prepare materials — clay, metal, wood, paints, or fabric", "Carve, mold, paint, or assemble products", "Quality check finished items for defects", "Pack products carefully for shipping or display", "Maintain tools and workspace", "Learn and practice new techniques to improve skills"},
		Skills: []string{"Specialized craft technique (metal, wood, painting, pottery, textile)", "Precision and attention to detail", "Creativity and design sense", "Knowledge of traditional Nepali motifs and designs", "Patience for time-intensive craft processes", "Basic business skills for self-employment", "Understanding of export quality standards", "Material knowledge and sourcing"},
		SalaryMin: 120000, SalaryMax: 600000, Difficulty: 3, FutureProof: 30,
		EducationReq: "Traditional apprenticeship with master artisan is the primary path. CTEVT programs in handicraft and applied arts. Training at craft development centers under the Federation of Handicraft Associations of Nepal (FHAN). Knowledge of traditional designs and techniques essential. Business skills for those selling directly.",
		Outlook: "Nepali handicrafts have steady demand from tourists and export markets. The industry faces challenges: competition from machine-made products, aging artisan population (young people not learning traditional skills), and changing market preferences. Artisans who adapt designs for modern tastes while maintaining traditional quality have better prospects. E-commerce (Etsy, Amazon) provides global market access.",
		Tags: []string{"other", "handicraft", "traditional", "artisan"},
		Resources: []resourceSeed{
			{Title: "Federation of Handicraft Associations of Nepal (FHAN)", URL: "https://www.handicraft.org.np", Description: "Umbrella organization for Nepal's handicraft sector"},
			{Title: "Merojob Handicraft Jobs", URL: "https://www.merojob.com", Description: "Find handicraft artisan jobs in Nepal"},
			{Title: "Traditional Nepali crafts documentation (YouTube)", URL: "https://www.youtube.com/results?search_query=traditional+nepali+handicraft+making"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn the basics of your chosen craft", Description: "Choose a handicraft specialty that interests you — metal statue making, thangka painting, wood carving, pottery, or textile crafts. Start learning the basics from family members if you come from an artisan family, or find a master artisan willing to teach you. The traditional method is apprenticeship — spending years learning from a master. Practice the fundamental techniques: for metal work — wax modeling, casting, finishing; for thangka — drawing proportions, mixing mineral paints, gold application; for wood carving — tool handling, design transfer, carving techniques. Be patient — handicraft skills take years to develop. Quality requires time and dedication.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Nepali metal statue making process (YouTube)", URL: "https://www.youtube.com/results?search_query=nepali+metal+statue+making+process"},
				{Title: "Thangka painting for beginners", URL: "https://www.youtube.com/results?search_query=thangka+painting+for+beginners"},
				{Title: "Wood carving techniques - Nepali style", URL: "https://www.youtube.com/results?search_query=nepali+wood+carving+techniques"},
			}},
			{StepNumber: 2, Title: "Get formal training and refine your skills", Description: "Enroll in craft training programs offered by FHAN, craft development centers, or CTEVT. Formal training teaches standardized techniques, quality control, and design principles. Learn about different craft types and materials. Study traditional Nepali designs — their symbolism, history, and cultural significance. Understanding the meaning behind designs makes your work more authentic and valuable. Learn to create your own designs while respecting traditional motifs. Develop speed without sacrificing quality. Master artisans can create pieces faster while maintaining exquisite quality. Speed comes with years of practice. Aim to produce pieces that are worthy of Nepal's reputation for fine craftsmanship.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "FHAN - Craft Development Training", URL: "https://www.handicraft.org.np"},
				{Title: "CTEVT Applied Arts Programs", URL: "https://www.ctevt.org.np"},
				{Title: "Nepali traditional motifs and their meanings", URL: "https://www.youtube.com/results?search_query=nepali+traditional+motifs+symbolism"},
			}},
			{StepNumber: 3, Title: "Produce quality pieces and build a portfolio", Description: "Create a portfolio of your best work — photographs of finished pieces documenting different types and designs. Focus on quality over quantity. A single excellent piece is worth more than ten average ones. Nepal's reputation for handicrafts depends on quality. Learn about export quality standards — dimensions, finish, packaging requirements. Show your work at local craft exhibitions, tourist areas, and craft shops. Build relationships with handicraft exporters and shop owners. Get feedback from experienced artisans and customers. Use feedback to improve your work. Word of mouth in the handicraft community is important. Good craftspeople are known and respected.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Nepal handicraft exhibition and trade fairs", URL: "https://www.handicraft.org.np"},
				{Title: "Photographing handicraft products for portfolio", URL: "https://www.youtube.com/results?search_query=product+photography+for+handicrafts"},
				{Title: "Export quality standards for Nepali crafts", URL: "https://www.fhan.org.np"},
			}},
			{StepNumber: 4, Title: "Learn business and marketing skills", Description: "If you want to earn from your craft, you need business skills. Learn to price your work properly — account for materials, time, overhead, and profit. Understand the handicraft market — what sells to tourists, what sells for export, what sells locally. Learn about export procedures if targeting international markets. Develop your brand — a distinct style that customers recognize. Create a social media presence (Instagram, Facebook) to showcase your work. Etsy and other online platforms connect artisans with global customers. Many Nepali artisans now sell directly through social media, bypassing middlemen and earning more. Business skills are as important as craft skills for financial success.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Pricing your handicraft products (YouTube)", URL: "https://www.youtube.com/results?search_query=how+to+price+handicrafts+for+sale"},
				{Title: "Selling Nepali handicrafts on Etsy guide", URL: "https://www.youtube.com/results?search_query=etsy+seller+guide+nepal+handicraft"},
				{Title: "Export procedures for handicrafts from Nepal", URL: "https://www.handicraft.org.np"},
			}},
			{StepNumber: 5, Title: "Master your craft and innovate designs", Description: "Master craftspeople are those who have spent 10+ years perfecting their technique. They can create complex pieces with exceptional quality. Master artisans also innovate — creating new designs that appeal to modern tastes while maintaining traditional quality and essence. Study market trends — what designs are popular with tourists, what international buyers want, what interior decorators look for. Adapt your craft to new applications — thangka designs on fashion accessories, wood carving on modern furniture, metal crafts for home decor. Innovation keeps the craft tradition alive and commercially viable. Master artisans are respected as living treasures of Nepali culture.",
			Duration: "3-5 years", Links: []roadmapLink{
				{Title: "Adapting traditional crafts for modern markets (YouTube)", URL: "https://www.youtube.com/results?search_query=modern+traditional+craft+design+nepal"},
				{Title: "Market trends for Nepali handicrafts", URL: "https://www.handicraft.org.np"},
				{Title: "Master artisan stories from Nepal", URL: "https://www.youtube.com/results?search_query=nepali+master+artisan+interview"},
			}},
			{StepNumber: 6, Title: "Train next generation and preserve craft heritage", Description: "Master artisans have a responsibility to teach the next generation. Take apprentices, teach at vocational training institutes, or document your techniques for future generations. Nepal's traditional crafts face the risk of dying out as young people choose other careers. Your teaching helps preserve cultural heritage. Some established artisans form cooperatives or workshops that employ multiple craftspeople, creating sustainable craft businesses. Others become craft entrepreneurs, exporting products and representing Nepali handicrafts internationally. Handicraft is not just a job — it is carrying forward centuries of Nepali artistic tradition. Your skilled hands create objects of beauty and cultural significance that can last for generations.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Handicraft cooperatives and artisan groups Nepal", URL: "https://www.handicraft.org.np"},
				{Title: "Teaching and preserving traditional crafts (YouTube)", URL: "https://www.youtube.com/results?search_query=preserving+traditional+crafts+nepal"},
				{Title: "Craft entrepreneurship and export business", URL: "https://www.youtube.com/results?search_query=handicraft+export+business+nepal"},
			}},
		},
	}
}

func cateringService() careerSeed {
	return careerSeed{
		CategoryName: "Other Essential Nepal Careers", CategorySlug: "other-essential", CategoryIcon: "🍽️",
		Title: "Catering Service Owner", Slug: "catering-service",
		Summary: "Caterers prepare and deliver food for events — weddings, parties, corporate events, and festivals. They manage menus, cooking, staff, and event food service.",
		Description: "A catering service provides food for events — weddings (the largest market in Nepal), parties, corporate events, religious ceremonies, festivals, and large gatherings. Catering in Nepal ranges from small home-based businesses serving neighborhood events to large commercial catering companies serving hundreds of guests. Nepali cuisine offers rich variety — dal bhat, curry dishes, momo, chowmein, traditional Newari food, and festival specialties. Caterers must understand menu planning, quantity cooking, food safety, presentation, and event logistics. The wedding season in Nepal (mostly November-February and April-June) is the peak business period. Successful caterers build reputations for delicious food, reliable service, and beautiful presentation. Catering is a scalable business — starting small with family help and growing to a full commercial operation.",
		DailyTasks: []string{"Plan menus with clients for their events", "Purchase fresh ingredients — vegetables, meat, spices, rice", "Prepare and cook food in large quantities", "Pack and transport food to event venues", "Set up food service — buffet, serving stations, or plated meals", "Manage catering staff during events", "Clean up after events and maintain kitchen hygiene"},
		Skills: []string{"Cooking expertise in Nepali cuisine (and other cuisines)", "Quantity cooking for large groups", "Menu planning and costing", "Food safety and hygiene standards", "Event logistics and time management", "Staff management and training", "Customer service and client communication", "Budgeting and pricing"},
		SalaryMin: 180000, SalaryMax: 900000, Difficulty: 2, FutureProof: 50,
		EducationReq: "SEE/+2 pass. Food safety training and health permit. Culinary training from CTEVT or hospitality institutes (NCHMCT). Experience in cooking and event coordination. Business registration and kitchen inspection certification. Knowledge of Nepali food traditions essential.",
		Outlook: "Catering demand is driven by weddings, which are elaborate affairs in Nepal. Corporate events and parties are growing segments. The industry is competitive — quality and reputation determine success. Home-based catering is a low-cost entry point. Full-service catering with decoration and staff commands higher prices. Food trends (healthy options, international cuisines) create opportunities for innovation.",
		Tags: []string{"other", "catering", "food", "entrepreneurship"},
		Resources: []resourceSeed{
			{Title: "NCHMCT (Nepal Academy of Tourism and Hotel Management)", URL: "https://www.nchmct.edu.np", Description: "Hospitality and catering education in Nepal"},
			{Title: "Department of Food Technology and Quality Control Nepal", URL: "https://www.dftqc.gov.np", Description: "Food safety regulations and hygiene standards"},
			{Title: "Merojob Catering Jobs", URL: "https://www.merojob.com", Description: "Find catering jobs and opportunities in Nepal"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Develop your cooking skills and signature dishes", Description: "Master Nepali cooking — dal bhat, curry varieties, momo, chowmein, sekwa, traditional Newari dishes, and festival foods. Practice cooking in large quantities — catering is about feeding many people. Learn to cook consistently — the same dish should taste the same every time. Develop 3-5 signature dishes that customers rave about. Your unique recipes and cooking style become your brand. Learn food presentation — how food looks is almost as important as how it tastes. Catering is visual — beautifully presented food attracts clients. Also learn about other cuisines — Indian, Chinese, Continental — to offer variety. If you are a good cook who people compliment, catering could be your path.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Nepali cooking masterclass (YouTube)", URL: "https://www.youtube.com/results?search_query=nepali+cuisine+recipes+for+beginners"},
				{Title: "Quantity cooking tips for large groups", URL: "https://www.youtube.com/results?search_query=bulk+cooking+tips+for+beginners"},
				{Title: "Food presentation and plating techniques", URL: "https://www.youtube.com/results?search_query=food+presentation+plating+techniques+guide"},
			}},
			{StepNumber: 2, Title: "Get food safety training and business registration", Description: "Obtain food safety and hygiene certification from the Department of Food Technology and Quality Control. Learn about: proper food storage, temperature control, cross-contamination prevention, personal hygiene, and kitchen sanitation. Catering businesses must meet health department standards. Register your catering business with the local municipality. Get a health permit for your kitchen. Learn about food licensing requirements in Nepal. Consider taking culinary courses at NCHMCT or other hospitality institutes to strengthen your knowledge. Formal training in food safety and business operations protects your customers and your business.",
			Duration: "2-4 months", Links: []roadmapLink{
				{Title: "Food safety training Nepal - DFTQC", URL: "https://www.dftqc.gov.np"},
				{Title: "Food business registration in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=food+business+registration+nepal+process"},
				{Title: "Kitchen hygiene and sanitation standards", URL: "https://www.youtube.com/results?search_query=kitchen+hygiene+standards+guide"},
			}},
			{StepNumber: 3, Title: "Start with small events and build reputation", Description: "Begin by catering small events — family gatherings, birthday parties, small pujas. Start from your home kitchen. Keep costs low initially — use existing kitchen equipment, buy ingredients from local markets, recruit family members as helpers. Focus on delivering excellent food and service. Word of mouth is the most powerful marketing for caterers. At every event, do your best. Satisfied customers will recommend you for bigger events. Build a simple menu card with prices. Create photo albums of your catering setups. Collect testimonials from clients. Your first 20 events are for building reputation, not maximizing profit.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Starting a home-based catering business (YouTube)", URL: "https://www.youtube.com/results?search_query=start+home+catering+business+nepal"},
				{Title: "How to price your catering services", URL: "https://www.youtube.com/results?search_query=catering+pricing+guide+nepal"},
				{Title: "Marketing your catering business locally", URL: "https://www.youtube.com/results?search_query=catering+business+marketing+tips"},
			}},
			{StepNumber: 4, Title: "Build your catering team and equipment", Description: "As you get more orders, expand your capacity. Hire reliable cooks and helpers. Train them in your cooking style and quality standards. Invest in catering equipment: large pots and pans, commercial stove (if budget allows), chafing dishes for buffets, serving utensils, transport containers, and possibly a delivery vehicle. Establish relationships with reliable ingredient suppliers — vegetables, meat, spices, rice. Consistent ingredient quality ensures consistent food quality. Create a system for managing events: client consultation, menu planning, ingredient procurement, cooking schedule, transport, setup, service, and clean-up. Organized systems enable you to handle larger events.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Catering equipment essentials for starting (YouTube)", URL: "https://www.youtube.com/results?search_query=catering+business+equipment+essentials"},
				{Title: "Building and training a catering team", URL: "https://www.youtube.com/results?search_query=how+to+build+a+catering+team"},
				{Title: "Supplier management for catering business", URL: "https://www.youtube.com/results?search_query=ingredient+sourcing+for+catering+business"},
			}},
			{StepNumber: 5, Title: "Specialize in wedding or corporate catering", Description: "Wedding catering is the biggest market in Nepal. Weddings involve multiple meals over 1-3 days, elaborate menus, and large guest counts (200-1000+ people). Wedding catering requires extensive planning, large teams, and significant equipment. Corporate catering (office events, conferences, company parties) is more regular and predictable. Some caterers specialize in specific cuisines: traditional Nepali wedding feasts, Newari cuisine, Indian catering, or continental options. Develop wedding packages at different price points. Build relationships with wedding venues, event planners, and decorators who can refer clients to you. Wedding caterers with good reputations have full booking calendars during wedding season.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "Wedding catering business guide Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=wedding+catering+business+nepal"},
				{Title: "Corporate catering - building regular contracts", URL: "https://www.youtube.com/results?search_query=corporate+catering+business+guide"},
				{Title: "Nepali wedding feast menu planning", URL: "https://www.youtube.com/results?search_query=nepali+wedding+menu+planning"},
			}},
			{StepNumber: 6, Title: "Expand your catering business or open a venue", Description: "Successful caterers expand in several ways: open a commercial kitchen and catering hall (event space with catering), start a restaurant that also does catering, operate multiple catering teams for simultaneous events, offer related services (decoration, tent rental, entertainment coordination), or develop a franchise model. The food business is demanding but rewarding. A successful catering business in Nepal can generate substantial income during peak seasons. More importantly, your food becomes part of people's most important celebrations. Your cooking feeds weddings, festivals, and gatherings where families and communities come together. Food is at the heart of every celebration.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Scaling a catering business in Nepal", URL: "https://www.youtube.com/results?search_query=scale+catering+business+nepal"},
				{Title: "Opening a banquet hall with catering", URL: "https://www.merojob.com"},
				{Title: "Catering business diversification ideas", URL: "https://www.youtube.com/results?search_query=catering+business+diversification"},
			}},
		},
	}
}

func pashminaBusiness() careerSeed {
	return careerSeed{
		CategoryName: "Other Essential Nepal Careers", CategorySlug: "other-essential", CategoryIcon: "🧣",
		Title: "Pashmina Business Owner", Slug: "pashmina-business",
		Summary: "Pashmina business owners produce, manufacture, or sell pashmina shawls, scarves, and products made from high-quality cashmere wool, a major Nepali export.",
		Description: "A pashmina business involves the production, manufacturing, or sale of pashmina products — primarily shawls, scarves, stoles, and blankets made from cashmere wool (chyangra pashmina from Himalayan mountain goats). Pashmina is one of Nepal's most famous exports, prized worldwide for its softness, warmth, and quality. The business ranges from small shops in tourist areas (Thamel, Pokhara) to large manufacturing and export operations. Pashmina business owners must understand: raw material sourcing (cashmere wool from the Himalayas), production processes (spinning, weaving, dyeing, finishing), quality control (authenticity, thread count, softness), and marketing (branding, export documentation, retail sales). Nepal's pashmina industry faces competition from machine-made imitations — authentic handmade pashmina commands premium prices. The business requires knowledge of textiles, quality assessment, and international trade.",
		DailyTasks: []string{"Source quality cashmere wool and other materials", "Supervise or coordinate pashmina production (weaving, dyeing)", "Inspect finished products for quality and authenticity", "Manage inventory of raw materials and finished products", "Serve retail customers in shop or showroom", "Process wholesale and export orders", "Market products through social media, trade shows, and wholesale channels"},
		Skills: []string{"Knowledge of cashmere wool types and quality grades", "Understanding of pashmina production processes", "Quality assessment and authentication skills", "Product design and color sense", "Retail sales and customer service", "Export documentation and international trade knowledge", "Inventory management and accounting", "Marketing and brand building"},
		SalaryMin: 240000, SalaryMax: 2000000, Difficulty: 3, FutureProof: 40,
		EducationReq: "Bachelor's in Business or Textile Management helpful but not required. Knowledge of pashmina industry through family business or apprenticeship. Training in textile quality assessment. Export-import training from Trade and Export Promotion Center. Business registration with Department of Cottage and Small Industries.",
		Outlook: "Nepali pashmina is globally recognized for quality. The industry faces challenges: competition from machine-made imitations (often labeled as pashmina), fluctuating raw material prices, and changing fashion trends. Authentic, high-quality pashmina maintains demand in international markets. E-commerce provides direct-to-consumer sales opportunities. Brand-building and certification (hallmark, authenticity guarantee) add value.",
		Tags: []string{"other", "pashmina", "export", "textile", "entrepreneurship"},
		Resources: []resourceSeed{
			{Title: "Trade and Export Promotion Center Nepal", URL: "https://www.tepc.gov.np", Description: "Export information and support for Nepali products including pashmina"},
			{Title: "Central Bureau of Statistics - Pashmina Trade Data", URL: "https://www.cbs.gov.np", Description: "Nepal trade statistics for pashmina industry"},
			{Title: "Nepal Pashmina Manufacturers Association", URL: "https://www.pashminanepal.com", Description: "Industry body for pashmina manufacturers and exporters"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn about pashmina and the textile industry", Description: "Study everything about pashmina: what makes authentic pashmina (cashmere wool from Chyangra goats), the production process (spinning, weaving, dyeing, finishing), quality grades, and how to distinguish real pashmina from imitations. Learn about Nepal's pashmina industry — major production areas (Kathmandu Valley, Pokhara), supply chain (wool from Himalayas, processing in Kathmandu), and export markets (USA, Europe, Japan, China). Visit pashmina factories and shops. Talk to people in the industry. Understand the challenges — imitation products, price competition, and changing fashion trends. Knowledge of the product and industry is essential before investing.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "Authentic vs fake pashmina - how to tell the difference (YouTube)", URL: "https://www.youtube.com/results?search_query=real+vs+fake+pashmina+identification"},
				{Title: "Pashmina production process in Nepal", URL: "https://www.youtube.com/results?search_query=pashmina+making+process+nepal"},
				{Title: "Nepal pashmina industry overview", URL: "https://www.pashminanepal.com"},
			}},
			{StepNumber: 2, Title: "Get hands-on experience in the pashmina trade", Description: "Work in a pashmina shop, factory, or export company to learn the practical aspects. Learn to assess quality — feel the fabric, check the weave, verify thread count, and identify different quality grades. Learn about different product types — shawls, scarves, stoles, blankets, sweaters, and how each is made. Understand pricing — raw material costs, production costs, wholesale pricing, and retail markups. Learn customer preferences — what designs sell, what colors are popular, what price points work for different markets. Build relationships with manufacturers, weavers, and suppliers. Experience in the trade is the best education for running your own business.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Pashmina quality assessment guide (YouTube)", URL: "https://www.youtube.com/results?search_query=pashmina+quality+check+guide"},
				{Title: "Find pashmina industry jobs (Merojob)", URL: "https://www.merojob.com"},
				{Title: "Understanding pashmina pricing and margins", URL: "https://www.youtube.com/results?search_query=pashmina+pricing+wholesale+retail"},
			}},
			{StepNumber: 3, Title: "Start with a small retail or online operation", Description: "Begin small: open a small shop in a tourist area (Thamel, Pokhara, or near a major temple), or start selling online through Etsy, Amazon, or your own website. Source products from established manufacturers. Carry 20-50 different designs initially. Focus on quality and authenticity — provide certificates of authenticity for genuine pashmina. Build your brand around quality, authenticity, and fair trade. Develop a story about your pashmina — where it comes from, how it is made, the artisans behind it. Customers pay premium for authentic pashmina with a credible story. Use social media (Instagram, Facebook) to showcase your products. Good product photography is essential for online sales.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Starting a pashmina business in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=start+pashmina+business+nepal"},
				{Title: "Selling pashmina on Etsy - guide", URL: "https://www.youtube.com/results?search_query=etsy+pashmina+seller+guide"},
				{Title: "Product photography for pashmina and textiles", URL: "https://www.youtube.com/results?search_query=product+photography+for+pashmina"},
			}},
			{StepNumber: 4, Title: "Develop your own products and brand", Description: "Move from selling generic products to developing your own branded line. Work with weavers and manufacturers to create exclusive designs — unique patterns, colors, and product types. Register your brand and create professional packaging. Develop a signature style that customers recognize. Build relationships with international buyers through trade fairs (Nepal International Trade Fair, handicraft exhibitions in Kathmandu). Consider getting your pashmina certified for authenticity (hallmarking, organic certification). A strong brand with authentic, quality products commands premium prices. Direct relationships with international buyers (wholesale) provide volume and regular orders.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Branding your pashmina products (YouTube)", URL: "https://www.youtube.com/results?search_query=pashmina+branding+tips"},
				{Title: "Nepal International Trade Fair participation", URL: "https://www.tepc.gov.np"},
				{Title: "Pashmina certification and hallmarking", URL: "https://www.pashminanepal.com"},
			}},
			{StepNumber: 5, Title: "Expand into manufacturing or export", Description: "Expand from retail into manufacturing or export. Manufacturing involves setting up a production unit with weaving looms, dyeing facilities, and finishing equipment. Employ skilled weavers and artisans. Export requires understanding of international trade documentation, customs procedures, and international shipping. Develop relationships with international distributors and retailers. Attend international trade fairs (India, Hong Kong, USA) to find buyers. Export volumes require capital for raw material purchase, production, and shipping. The profit margins on export are narrower but volumes are much larger. Many successful pashmina businesses combine retail (high margin, lower volume) with export (lower margin, higher volume) for balanced growth.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "Pashmina manufacturing setup guide (YouTube)", URL: "https://www.youtube.com/results?search_query=pashmina+manufacturing+unit+nepal"},
				{Title: "Export documentation for pashmina from Nepal", URL: "https://www.tepc.gov.np"},
				{Title: "International trade fairs for textile and fashion", URL: "https://www.youtube.com/results?search_query=textile+trade+fair+participation+nepal"},
			}},
			{StepNumber: 6, Title: "Build a pashmina brand recognized globally", Description: "Established pashmina businesses build globally recognized brands. This requires consistent quality, professional marketing, excellent customer service, and sometimes celebrity/press coverage. Some successful pashmina entrepreneurs expand into related products — cashmere clothing, home textiles (cashmere throws, cushions), and other handloom products. Others diversify into other Nepali handicrafts. The pashmina industry supports many Nepali families — from goat herders in the Himalayas to weavers in Kathmandu Valley. A successful pashmina business is not just profitable — it helps preserve traditional craftsmanship and provides livelihoods for artisans. Your pashmina brand carries the legacy of Nepali textile tradition.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Nepali pashmina brands going global", URL: "https://www.youtube.com/results?search_query=nepali+pashmina+brand+global+success"},
				{Title: "Brand building for Nepali products (YouTube)", URL: "https://www.youtube.com/results?search_query=brand+building+nepali+handicraft"},
				{Title: "Sustainable and ethical pashmina business", URL: "https://www.pashminanepal.com"},
			}},
		},
	}
}

func carpetManufacturer() careerSeed {
	return careerSeed{
		CategoryName: "Other Essential Nepal Careers", CategorySlug: "other-essential", CategoryIcon: "🕌",
		Title: "Carpet Manufacturer", Slug: "carpet-manufacturer",
		Summary: "Carpet manufacturers produce hand-knotted and hand-tufted wool carpets — a major Nepali export industry known worldwide for quality and design.",
		Description: "A carpet manufacturer produces hand-knotted and hand-tufted wool carpets — one of Nepal's largest export industries. Nepali carpets are famous for their quality, design, and use of New Zealand and Tibetan wool. The industry employs thousands of skilled weavers and supports many related businesses (wool processing, dyeing, washing, finishing). Carpet manufacturers handle: raw material procurement (wool, silk, dyes), production management (design, weaving, quality control), finishing (washing, cutting, binding), and sales (export, wholesale). The Nepal carpet industry exports to over 40 countries with major markets in the USA, Europe, Australia, and Asia. The business requires significant capital, knowledge of textile manufacturing, management of large teams of weavers, and understanding of international trade. Carpet manufacturing is deeply connected to Nepal's economy and Tibetan refugee community (Tibetan refugees in Nepal are traditionally skilled carpet weavers).",
		DailyTasks: []string{"Source wool, silk, and other raw materials", "Create or select carpet designs and color schemes", "Supervise weaving operations and quality control", "Manage washing, cutting, and finishing processes", "Inspect finished carpets for quality standards", "Coordinate with export agents and international buyers", "Manage production schedules and inventory"},
		Skills: []string{"Knowledge of carpet production processes (design, dyeing, weaving, finishing)", "Wool and silk quality assessment", "Design sense and understanding of market trends", "Production management and quality control", "Export documentation and international trade", "Team management of weavers and production staff", "Financial management for manufacturing operations", "Knowledge of international quality standards (CFL, label certifications)"},
		SalaryMin: 300000, SalaryMax: 3000000, Difficulty: 4, FutureProof: 35,
		EducationReq: "Bachelor's in Business Management or Textile Engineering helpful. Extensive knowledge of carpet industry gained through experience. Training from Carpet and Handicraft Development Center. Export-import training. Nepali/Tibetan language helpful for managing weavers. Business registration with Department of Cottage and Small Industries.",
		Outlook: "Nepali carpet industry is a major export earner but faces challenges: competition from machine-made carpets (India, China, Turkey), labor shortages (young people not choosing weaving), raw material price fluctuations, and changing interior design trends. Manufacturers who adapt to market trends (contemporary designs, eco-friendly materials, custom sizes) have better prospects. Certification (GoodWeave, ensuring no child labor) is important for export markets. The industry has stabilized after decline in the 2000s.",
		Tags: []string{"other", "carpet", "export", "manufacturing", "textile"},
		Resources: []resourceSeed{
			{Title: "The Carpet and Handicraft Development Center (CHDC)", URL: "https://www.carpetnepal.com", Description: "Carpet industry development and export promotion in Nepal"},
			{Title: "GoodWeave Nepal", URL: "https://www.goodweave.org.np", Description: "Child labor free carpet certification and industry ethics"},
			{Title: "Trade and Export Promotion Center Nepal", URL: "https://www.tepc.gov.np", Description: "Export information and support for Nepali carpet industry"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Learn the carpet industry from the ground floor", Description: "Start by working in a carpet factory or showroom. Learn the complete production process: wool selection and grading, dyeing (natural and chemical dyes), design creation (traditional Tibetan, contemporary, custom), weaving (hand-knotting techniques, knots per square inch - KPSI), washing and chemical treatment, finishing (cutting, binding, edge finishing), and quality inspection. Understand different carpet types: Tibetan (hand-knotted with high knot density), Newari, and woolen tufted. Learn about quality factors: wool quality, knot density, design precision, color fastness, and finishing quality. The carpet industry is complex — a thorough understanding of the production process is essential before managing a manufacturing business.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Nepali carpet making process (YouTube)", URL: "https://www.youtube.com/results?search_query=nepali+carpet+making+process+documentary"},
				{Title: "Understanding hand-knotted carpet quality", URL: "https://www.youtube.com/results?search_query=hand+knotted+carpet+quality+guide"},
				{Title: "CHDC - Carpet industry information", URL: "https://www.carpetnepal.com"},
			}},
			{StepNumber: 2, Title: "Learn production management and quality control", Description: "Study production management: scheduling weaving orders, managing raw material inventory, maintaining quality standards, and ensuring on-time delivery. Learn to calculate production costs — wool cost, labor cost, dyeing, finishing, overhead — to price carpets profitably. Understand international quality standards: GoodWeave certification (ensuring no child labor), CFL (carpet labeling) standards, and customer-specific quality requirements. Learn about export procedures: packing, shipping documentation, customs clearance. American and European buyers have strict quality requirements. Build relationships with wool suppliers (New Zealand wool imported through India, Tibetan wool from Himalayan regions) and dye suppliers.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "Carpet quality control standards (YouTube)", URL: "https://www.youtube.com/results?search_query=carpet+quality+control+standards+nepal"},
				{Title: "GoodWeave certification for Nepali carpets", URL: "https://www.goodweave.org.np"},
				{Title: "Carpet production cost calculation", URL: "https://www.youtube.com/results?search_query=carpet+manufacturing+cost+calculation"},
			}},
			{StepNumber: 3, Title: "Work in carpet sales or export", Description: "Work in carpet sales — either in a showroom (tourist or export-oriented) or as an export assistant. Learn what customers want — design preferences (traditional vs contemporary), size requirements, quality expectations, price points. Different markets prefer different styles: American buyers prefer contemporary designs in neutral colors, European buyers may prefer traditional Tibetan designs, and Australian buyers look for natural undyed wool carpets. Learn to negotiate with international buyers. Understand trade terms (FOB, CIF), payment methods (letter of credit, wire transfer), and shipping logistics. Building relationships with international buyers is the key to carpet export success. The carpet industry in Nepal operates heavily on trust and relationships.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Carpet export business guide (YouTube)", URL: "https://www.youtube.com/results?search_query=carpet+export+business+nepal"},
				{Title: "International market preferences for Nepali carpets", URL: "https://www.carpetnepal.com"},
				{Title: "Trade terms and payment in export business", URL: "https://www.tepc.gov.np"},
			}},
			{StepNumber: 4, Title: "Start your own carpet production unit", Description: "Starting a carpet manufacturing unit requires: factory space (in Kathmandu Valley or outside), weaving looms (frame looms for hand-knotting), trained weavers (the most critical resource — skilled weavers are increasingly scarce), raw materials (wool, silk, dyes), and finishing equipment (washing machines, drying area, shearing machines). Recruit skilled weavers — many are from Tibetan refugee families or traditional weaving communities in Nepal. Provide good working conditions and fair wages to retain skilled workers. Start small — 5-10 looms — and expand as orders grow. Build your reputation for quality and reliability. The first year is about establishing production capability and quality standards.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Setting up a carpet weaving factory (YouTube)", URL: "https://www.youtube.com/results?search_query=set+up+carpet+factory+nepal"},
				{Title: "Managing carpet weavers and production team", URL: "https://www.youtube.com/results?search_query=managing+carpet+weavers+nepal"},
				{Title: "CHDC support for new carpet manufacturers", URL: "https://www.carpetnepal.com"},
			}},
			{StepNumber: 5, Title: "Develop market presence and brand", Description: "Participate in international trade fairs for carpets and home furnishings (Domotex in Germany, Atlanta International Gift and Home Furnishings Market, Maison et Objet in Paris). Nepal's Carpet and Handicraft Development Center supports participation in international trade fairs. Develop a brand identity — company name, logo, catalog, website. Invest in professional photography of your carpets. Create a comprehensive catalog with designs, sizes, and quality options. Build relationships with interior designers, architects, and carpet dealers in target markets. Develop a website with e-commerce capability for direct sales. Many carpet manufacturers now sell directly to consumers through online platforms, offering custom sizes and designs.",
			Duration: "2-4 years", Links: []roadmapLink{
				{Title: "International carpet trade fairs participation", URL: "https://www.carpetnepal.com"},
				{Title: "Building a carpet brand online (YouTube)", URL: "https://www.youtube.com/results?search_query=online+carpet+brand+building"},
				{Title: "Marketing Nepali carpets to interior designers", URL: "https://www.youtube.com/results?search_query=nepali+carpets+interior+design+market"},
			}},
			{StepNumber: 6, Title: "Scale up and innovate in carpet manufacturing", Description: "Established manufacturers expand: increase loom capacity, diversify products (silk carpets, wool-silk blends, bamboo silk, custom sizes and shapes), open a showroom in Kathmandu for tourists and export buyers, develop direct relationships with retailers in export markets, or integrate vertically (wool processing, dyeing unit, finishing plant). Innovation opportunities: eco-friendly carpets (natural dyes, recycled materials), contemporary designs for modern interiors, custom design service (made-to-order based on customer specifications), and online configurators for custom carpets. The carpet industry is a proud part of Nepal's manufacturing sector. Your carpets adorn homes and offices worldwide, representing Nepali craftsmanship. Quality, ethics (GoodWeave certified), and design innovation are the keys to sustained success.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Scaling a carpet manufacturing business (YouTube)", URL: "https://www.youtube.com/results?search_query=scale+carpet+manufacturing+nepal"},
				{Title: "Eco-friendly and sustainable carpet production", URL: "https://www.goodweave.org.np"},
				{Title: "Future trends in handmade carpet industry", URL: "https://www.carpetnepal.com"},
			}},
		},
	}
}

func govtSchoolTeacher() careerSeed {
	return careerSeed{
		CategoryName: "Government Jobs",
		CategorySlug: "government-jobs",
		CategoryDesc: "Stable and respected careers in the public sector offering job security, pensions, and benefits.",
		CategoryIcon: "🏛️",
		Title:        "Government Secondary & Higher Level Teacher",
		Slug:         "government-secondary-higher-level-teacher",
		Summary:      "Government secondary and higher level teachers work in public schools across Nepal, teaching subjects like math, science, English, social studies, and more to students from grades 6 through 12.",
		Description:  "A government secondary and higher level teacher in Nepal is a permanent faculty member of a public school, appointed by the Nepal Public Service Commission (Lok Sewa Ayog) or the Teacher Service Commission. Secondary level teachers teach grades 6-10, while higher secondary teachers teach grades 11-12. This is considered one of the most stable and respected careers in Nepal — it offers a government salary scale (scaled with years of service), a guaranteed pension, annual leave, maternity/paternity benefits, and high social status. Teachers follow the national curriculum developed by the Curriculum Development Center (CDC). They prepare lesson plans, deliver instruction, assess students through exams and assignments, manage classrooms, and participate in school development activities. The job is demanding but rewarding — you shape the next generation. Government teachers in Nepal have strong union representation and collective bargaining power. Once you achieve permanent status (sthaye), your job is secure for life. Competition for government teaching posts is intense, especially for popular subjects and urban locations. The selection process includes a highly competitive written exam, interview, and demo teaching. Many teachers supplement their government salary with private tutoring, which is very common and lucrative in Nepal.",
		DailyTasks: []string{
			"Deliver lessons according to the national curriculum and school schedule",
			"Prepare detailed lesson plans with learning objectives and activities",
			"Create and grade exams, assignments, and homework",
			"Maintain student attendance, grade records, and progress reports",
			"Provide extra support to struggling students before or after school",
			"Participate in staff meetings, school events, and parent-teacher conferences",
			"Coordinate with other subject teachers on interdisciplinary projects",
		},
		Skills: []string{
			"In-depth knowledge of your subject area (Math, Science, English, Social Studies, etc.)",
			"Classroom management and student engagement",
			"Lesson planning and curriculum design",
			"Assessment and exam preparation",
			"Patience and empathy with adolescent learners",
			"Public speaking and clear communication in Nepali and English",
			"Basic educational technology skills",
			"Ability to prepare students for SEE, NEB, and scholarship exams",
		},
		SalaryMin:    400000,
		SalaryMax:    900000,
		SalaryCurrency: "NPR",
		SalaryPeriod: "yearly",
		Difficulty:   3,
		FutureProof:  85,
		EducationReq: "For secondary level (grades 6-10): Bachelor's degree in Education (B.Ed.) with major in relevant subject, or a Bachelor's degree in a subject plus a one-year B.Ed. For higher secondary (grades 11-12): Master's degree (M.Ed. or M.A.) in the relevant subject. Must pass the Teacher Service Commission (TSC) exam for permanent appointment.",
		Outlook:      "Government teaching is one of the most stable careers in Nepal. Demand is consistently high, especially for math, science, English, and computer teachers. The government regularly opens new positions through the TSC. Rural areas have severe teacher shortages, so there are often more opportunities outside the Kathmandu Valley. The pension system ensures financial security after retirement. Competition is strong but the job security and benefits make it worth pursuing.",
		Tags:         []string{"government", "teaching", "stable", "pension", "education", "public-service"},
		Resources: []resourceSeed{
			{Title: "Teacher Service Commission Nepal", URL: "https://tsc.gov.np", Description: "Official TSC website — exam notices, syllabus, results, and teacher recruitment information"},
			{Title: "Curriculum Development Center Nepal", URL: "https://moecdc.gov.np", Description: "National curriculum, textbooks, and teacher guides for all subjects and levels"},
			{Title: "Lok Sewa Ayog Nepal", URL: "https://psc.gov.np", Description: "Public Service Commission — government job notifications, exam schedules, and results"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Choose your teaching subject and level", Description: "Decide which subject you want to teach (Mathematics, Science, English, Nepali, Social Studies, etc.) and which level (secondary grades 6-10 or higher secondary grades 11-12). Your choice determines your degree path and the TSC exam you will take. Pick a subject you are genuinely passionate about — you will spend decades teaching it. Math, Science, and English teachers are in highest demand. Consider your own academic strengths. Talk to current government teachers about their experience. Visit a local public school and observe classes. Understanding the reality of teaching in Nepal helps you make an informed decision.",
			Duration: "1-3 months", Links: []roadmapLink{
				{Title: "Teacher Service Commission - Subject Requirements", URL: "https://tsc.gov.np"},
				{Title: "Nepal Government Teacher Career Guide (YouTube)", URL: "https://www.youtube.com/results?search_query=government+teacher+career+nepal"},
				{Title: "Teaching Subjects in Demand in Nepal", URL: "https://www.merojob.com"},
			}},
			{StepNumber: 2, Title: "Earn your required degree", Description: "For secondary level: complete a Bachelor of Education (B.Ed.) with a major in your chosen subject. If you already have a Bachelor's in another subject, complete a one-year B.Ed. program. For higher secondary: earn a Master's degree (M.Ed. or M.A.) in your subject area. Many universities in Nepal offer these programs — Tribhuvan University, Kathmandu University, Purbanchal University, and Pokhara University. Consider distance learning options if you need to work while studying. Your degree must be recognized by the TSC. Maintain good grades — TSC exam results may consider your academic performance.",
			Duration: "1-4 years", Links: []roadmapLink{
				{Title: "TU Central Department of Education", URL: "https://tribhuvan-university.edu.np"},
				{Title: "Kathmandu University School of Education", URL: "https://soe.ku.edu.np"},
				{Title: "Open and Distance Learning Education in Nepal", URL: "https://www.odl.tu.edu.np"},
			}},
			{StepNumber: 3, Title: "Prepare for and pass the TSC (Teacher Service Commission) exam", Description: "The TSC teacher selection exam is highly competitive. It consists of a written exam (subject knowledge, teaching methodology, and general knowledge) followed by an interview and demo teaching. Start preparing at least 6-12 months before the exam. Study the TSC syllabus for your subject thoroughly. Solve past exam questions. Consider joining a TSC preparation class in your city. Many coaching centers in Kathmandu, Pokhara, and major cities offer specialized TSC preparation. The competition is fierce — sometimes thousands of applicants for a few dozen positions. Consistent and dedicated preparation is essential. Your exam score determines your ranking and chances of getting a posting.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "TSC Exam Syllabus and Past Questions", URL: "https://tsc.gov.np"},
				{Title: "TSC Preparation Classes in Nepal (YouTube)", URL: "https://www.youtube.com/results?search_query=TSC+preparation+nepal"},
				{Title: "Lok Sewa General Knowledge Preparation", URL: "https://www.youtube.com/results?search_query=lok+sewa+general+knowledge+nepal"},
			}},
			{StepNumber: 4, Title: "Clear the interview and demo teaching round", Description: "If you pass the written exam, you will be called for an interview and demonstration teaching (demo class). The interview panel includes senior educators and TSC officials. They assess your subject knowledge, communication skills, teaching aptitude, and personality. The demo teaching is crucial — you will teach a real lesson to a panel acting as students. Prepare engaging lesson plans, use the board effectively, ask questions, and show classroom management skills. Practice demo teaching with friends or family beforehand. Be confident but humble. Show that you love teaching, not just the government job benefits. Dress professionally and arrive early.",
			Duration: "1-2 months", Links: []roadmapLink{
				{Title: "Teacher Interview Tips (YouTube)", URL: "https://www.youtube.com/results?search_query=teacher+interview+tips+nepal"},
				{Title: "Demo Teaching Techniques (YouTube)", URL: "https://www.youtube.com/results?search_query=demo+teaching+techniques+for+teachers"},
				{Title: "How to Prepare a Lesson Plan for Demo", URL: "https://www.youtube.com/results?search_query=lesson+plan+for+demo+teaching"},
			}},
			{StepNumber: 5, Title: "Get appointed and complete probation", Description: "Selected candidates are appointed to a specific school based on their ranking and available vacancies. You will serve a probation period (usually 1-2 years) before becoming a permanent (sthaye) teacher. During probation, your performance is evaluated by the school principal and local education office. Complete all required paperwork and registration. Build good relationships with your principal, colleagues, students, and parents. Learn the school's culture and systems. Your probation performance determines your confirmation as a permanent government teacher. Once confirmed, you have job security, a pension, and eligibility for promotions and transfers.",
			Duration: "1-2 years", Links: []roadmapLink{
				{Title: "Teacher Probation and Confirmation Process (TSC)", URL: "https://tsc.gov.np"},
				{Title: "Nepal Government Teacher Service Conditions", URL: "https://moe.gov.np"},
				{Title: "New Teacher Survival Guide for Nepal", URL: "https://www.youtube.com/results?search_query=new+teacher+advice+nepal"},
			}},
			{StepNumber: 6, Title: "Pursue promotions and professional development", Description: "Government teachers can be promoted through grades (e.g., From teacher to senior teacher to vice principal to principal). Promotions are based on seniority, additional qualifications, and performance evaluations. Pursue a Master's degree or M.Phil./PhD to qualify for higher positions and salary grades. Attend teacher training programs organized by the Education Training Center (ETC) or NGOs like Room to Read, VSO, and UNICEF. Develop expertise in curriculum development, educational leadership, or special education. Some teachers become education officers, curriculum developers, or teacher trainers. Consider studying abroad for advanced degrees through scholarships. The more qualified you are, the faster you advance.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "Education Training Center Nepal", URL: "https://www.etc.gov.np"},
				{Title: "Teacher Professional Development Programs", URL: "https://moe.gov.np"},
				{Title: "Government Teacher Promotion Criteria Nepal", URL: "https://tsc.gov.np"},
			}},
		},
	}
}

func youtuber() careerSeed {
	return careerSeed{
		CategoryName: "Media & Entertainment",
		CategorySlug: "media-entertainment",
		CategoryDesc: "Careers in content creation, entertainment, and digital media production.",
		CategoryIcon: "🎥",
		Title:        "YouTuber",
		Slug:         "youtuber",
		Summary:      "A YouTuber creates and publishes video content on YouTube, building an audience through engaging videos on topics like education, entertainment, vlogging, gaming, music, and more.",
		Description:  "A YouTuber is a content creator who produces videos for the YouTube platform. YouTubers can create content on virtually any topic — educational tutorials, travel vlogs, music covers, comedy sketches, gaming streams, cooking shows, tech reviews, motivational talks, and Nepali cultural content. Successful YouTubers earn money through multiple revenue streams: YouTube advertising (AdSense), brand sponsorships, affiliate marketing, merchandise sales, crowdfunding (Patreon, Ko-fi), and paid consulting. Nepal's YouTube ecosystem has grown explosively — creators like Sisan Baniya, Aananda Koirala, Nishan Bhattarai, and many others have shown that YouTubing is a viable full-time career in Nepal. You can start with just a smartphone and an idea. Success requires consistency, patience, and understanding your audience. Most successful YouTubers took 1-3 years before earning meaningful income. The key is finding a niche you are passionate about and creating content that provides value — entertainment, education, or inspiration. YouTube is not a get-rich-quick path but a legitimate career for those who treat it seriously. Top Nepali YouTubers earn more than many traditional professionals.",
		DailyTasks: []string{
			"Brainstorm and research video ideas based on trends and audience interests",
			"Write scripts or outline key points for each video",
			"Film video content using camera or smartphone",
			"Edit videos — cut footage, add effects, music, and thumbnails",
			"Upload and optimize videos with titles, descriptions, tags, and thumbnails",
			"Engage with audience by replying to comments and community posts",
			"Plan content calendar, track analytics, and collaborate with other creators",
		},
		Skills: []string{
			"Video scripting and storytelling",
			"On-camera presence and communication",
			"Video shooting (framing, lighting, audio)",
			"Video editing (Premiere Pro, DaVinci Resolve, CapCut)",
			"YouTube SEO — titles, tags, descriptions, thumbnails",
			"Audience engagement and community management",
			"Social media promotion (Instagram, TikTok, Facebook)",
			"Basic graphic design for thumbnails (Canva, Photoshop)",
		},
		SalaryMin:    0,
		SalaryMax:    5000000,
		SalaryCurrency: "NPR",
		SalaryPeriod: "yearly",
		Difficulty:   4,
		FutureProof:  60,
		EducationReq: "No formal education required. You learn by doing. Skills in video editing, scripting, and on-camera presentation matter far more than degrees. Free resources: YouTube Creator Academy, online video editing tutorials. Some creators benefit from media or communication studies at college.",
		Outlook:      "YouTube and the creator economy are growing rapidly in Nepal. Internet penetration and smartphone usage are increasing every year. More Nepali viewers are consuming digital content than ever before. YouTube monetization is accessible once you reach 1,000 subscribers and 4,000 watch hours. Brand sponsorship deals are becoming more common in Nepal. However, the field is getting more competitive — standing out requires quality, consistency, and a unique angle. Successful YouTubers diversify income across multiple platforms. The creator economy is still young in Nepal — early movers have a significant advantage.",
		Tags:         []string{"media", "content-creation", "youtube", "entertainment", "digital", "entrepreneurship"},
		Resources: []resourceSeed{
			{Title: "YouTube Creator Academy", URL: "https://creatoracademy.youtube.com", Description: "Free official YouTube training on growing your channel, creating content, and monetization"},
			{Title: "Canva for Thumbnails", URL: "https://www.canva.com", Description: "Free design tool for creating eye-catching YouTube thumbnails and channel art"},
			{Title: "Nepali YouTube Creator Community (Facebook)", URL: "https://www.facebook.com", Description: "Join Nepali YouTuber groups for collaboration, tips, and local sponsorship opportunities"},
		},
		RoadmapSteps: []roadmapSeed{
			{StepNumber: 1, Title: "Find your niche and define your channel", Description: "Choose what your channel will be about. The most successful YouTube channels focus on a specific niche rather than being general. Possible niches for Nepal: educational content (math, science, English, coding tutorials), travel vlogs (exploring Nepal's 77 districts), food reviews (Nepali restaurants and street food), comedy and skits, music covers, tech reviews (Nepali language), motivational and self-improvement, farming and agriculture, personal finance and investing, or gaming. Pick a niche that aligns with your interests, knowledge, and personality. You will create hundreds of videos — you must genuinely enjoy the topic. Research existing Nepali channels in your niche. Find a unique angle or gap you can fill. Your channel name, branding, and content style should reflect your personality and niche.",
			Duration: "1-2 months", Links: []roadmapLink{
				{Title: "How to Find Your YouTube Niche (YouTube)", URL: "https://www.youtube.com/results?search_query=find+your+youtube+niche"},
				{Title: "Top Nepali YouTube Channels Analysis", URL: "https://www.youtube.com/results?search_query=top+nepali+youtubers"},
				{Title: "YouTube Creator Academy - Find Your Niche", URL: "https://creatoracademy.youtube.com/page/lesson/niche-discovery"},
			}},
			{StepNumber: 2, Title: "Learn basic video production with what you have", Description: "You do not need expensive equipment to start. A smartphone with a good camera (most mid-range phones are fine) and basic free editing software (CapCut, DaVinci Resolve) are enough. Focus on three things: decent lighting (natural light from a window works), clear audio (record in a quiet room, speak clearly), and stable footage (rest your phone on something). Learn basic video editing: cutting, adding text, background music, and transitions. Watch YouTube tutorials for your editing app. Your first videos will not be great. Do not wait until you have perfect equipment. Start with what you have. The sooner you start, the sooner you improve. Many successful Nepali YouTubers started with just a phone and basic editing.",
			Duration: "1-3 months", Links: []roadmapLink{
				{Title: "How to Start YouTube with Just a Phone", URL: "https://www.youtube.com/results?search_query=start+youtube+with+phone+nepal"},
				{Title: "CapCut Video Editing Tutorial (Nepali)", URL: "https://www.youtube.com/results?search_query=capcut+editing+tutorial+nepali"},
				{Title: "DaVinci Resolve Free Editing Course", URL: "https://www.youtube.com/results?search_query=davinci+resolve+beginner+tutorial"},
			}},
			{StepNumber: 3, Title: "Create and upload your first 20-30 videos consistently", Description: "Consistency is the most important factor for YouTube growth. Commit to uploading at least one video per week. Do not judge your early videos harshly — they will be rough. That is normal. Focus on improving one thing each video: better thumbnail, clearer audio, tighter editing, more engaging script. Use YouTube Studio to learn basic SEO: write descriptive titles, fill out descriptions, use relevant tags, and design custom thumbnails. Your first 20-30 videos are practice. The goal is to develop your style, improve your production skills, and start understanding what your audience likes. Every creator's early videos are cringeworthy. Push through and keep creating. Analyze your analytics to see which videos perform better and why.",
			Duration: "3-6 months", Links: []roadmapLink{
				{Title: "YouTube SEO for Beginners (Nepali)", URL: "https://www.youtube.com/results?search_query=youtube+seo+nepali"},
				{Title: "How to Design YouTube Thumbnails (Canva)", URL: "https://www.youtube.com/results?search_query=design+youtube+thumbnail+canva"},
				{Title: "YouTube Studio Analytics Guide", URL: "https://creatoracademy.youtube.com/page/lesson/analytics"},
			}},
			{StepNumber: 4, Title: "Build your audience and community", Description: "Engage with your viewers by responding to comments sincerely. Ask questions in your videos to encourage comments. Create community posts to stay connected between uploads. Collaborate with other Nepali creators in your niche — collaboration exposes both channels to new audiences. Promote your videos on social media (Instagram, TikTok, Facebook, Twitter). Share behind-the-scenes content and teasers. Consistency matters more than viral videos. A channel that grows slowly but steadily with an engaged community is more valuable than a one-hit-wonder. Focus on building a loyal audience that watches every video, not just chasing views. A small but dedicated audience is the foundation of a sustainable YouTube career.",
			Duration: "6-12 months", Links: []roadmapLink{
				{Title: "How to Grow Your YouTube Channel (YouTube)", URL: "https://www.youtube.com/results?search_query=grow+youtube+channel+from+0"},
				{Title: "Nepali YouTuber Collaboration Ideas", URL: "https://www.youtube.com/results?search_query=nepali+youTuber+collaboration"},
				{Title: "YouTube Community Building Strategies", URL: "https://creatoracademy.youtube.com/page/lesson/community"},
			}},
			{StepNumber: 5, Title: "Apply for YouTube Partner Program and monetize", Description: "Once you reach 1,000 subscribers and 4,000 watch hours in the past 12 months, apply for the YouTube Partner Program (YPP). Once approved, you can earn money through AdSense (ads shown on your videos). First earnings are typically small — do not quit your day job immediately. As your channel grows, diversify income: brand sponsorships (Nepali brands pay creators for promotion), affiliate marketing (earn commission promoting products), merchandise (t-shirts, mugs, books), crowdfunding (Patreon for exclusive content), and super chats during live streams. Multiple income streams are essential for financial stability as a creator. Many Nepali YouTubers also offer consulting, coaching, or paid appearances. Keep creating consistently even after monetization — your audience is your asset.",
			Duration: "1-3 years", Links: []roadmapLink{
				{Title: "YouTube Partner Program Requirements", URL: "https://support.google.com/youtube/answer/72857"},
				{Title: "How Nepali YouTubers Make Money", URL: "https://www.youtube.com/results?search_query=how+do+nepali+youtubers+make+money"},
				{Title: "Brand Sponsorship Guide for Creators", URL: "https://www.youtube.com/results?search_query=brand+sponsorship+for+youtubers"},
			}},
			{StepNumber: 6, Title: "Scale your channel into a media business", Description: "As your channel grows, treat it as a business. Consider hiring a video editor, thumbnail designer, or social media manager. Reinvest earnings into better equipment (camera, microphone, lighting). Diversify across platforms — start a podcast, build a TikTok following, create an Instagram presence, launch a newsletter. Consider creating digital products (courses, ebooks) or launching a membership community. Some successful YouTubers expand into offline businesses — events, workshops, merchandise stores. The most successful creators build a personal brand that extends far beyond YouTube. A YouTube channel is a powerful platform that can launch many other career opportunities. Top Nepali creators are now earning incomes comparable to doctors, engineers, and business owners.",
			Duration: "ongoing", Links: []roadmapLink{
				{Title: "From YouTuber to Media Business (YouTube)", URL: "https://www.youtube.com/results?search_query=scale+youtube+channel+into+business"},
				{Title: "Building a Personal Brand as a Creator", URL: "https://www.youtube.com/results?search_query=personal+branding+for+creators"},
				{Title: "Nepali Digital Creator Economy Trends", URL: "https://www.youtube.com/results?search_query=nepal+creator+economy+2026"},
			}},
		},
	}
}

func seedAdminUser(db *sqlx.DB) {
	emails := []string{
		"siddharthadhakal3722@gmail.com",
		"siddharthadhakall3722@gmail.com",
	}

	for _, email := range emails {
		var count int
		db.Get(&count, `SELECT COUNT(*) FROM users WHERE email = $1`, email)
		if count > 0 {
			db.Exec(`UPDATE users SET is_admin = true WHERE email = $1`, email)
			log.Printf("admin user %s already exists, set admin flag", email)
			continue
		}

		hash, err := bcrypt.GenerateFromPassword([]byte("Balakotalu77"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("failed to hash admin password: %v", err)
		}

		_, err = db.Exec(
			`INSERT INTO users (email, password_hash, name, is_admin) VALUES ($1, $2, $3, true)`,
			email, string(hash), "admin",
		)
		if err != nil {
			log.Fatalf("failed to create admin user %s: %v", email, err)
		}

		log.Printf("admin user %s seeded", email)
	}
}
