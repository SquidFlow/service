import Link from "next/link"

export default function Footer() {
    return (
        <footer className="bg-background border-t border-border">
            <div className="max-w-7xl mx-auto py-4 px-4 sm:px-6 lg:px-8">
                <div className="flex justify-between items-center">
                    <p className="text-sm text-muted-foreground">&copy; 2023 H4 Platform. All rights reserved.</p>
                    <nav>
                        <ul className="flex space-x-4">
                            <li><Link href="/privacy" className="text-sm text-muted-foreground hover:text-primary">Privacy Policy</Link></li>
                            <li><Link href="/terms" className="text-sm text-muted-foreground hover:text-primary">Terms of Service</Link></li>
                        </ul>
                    </nav>
                </div>
            </div>
        </footer>
    )
}