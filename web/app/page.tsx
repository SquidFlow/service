"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import Dog3D from './components/dog3d'
import Header from './components/header'
import Footer from './components/footer'
import { Zap, Shield, Package, Workflow, Database, Cloud, Lock } from 'lucide-react'
import Image from 'next/image'
import { Button } from "@/components/ui/button"
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion"
import { useRouter } from 'next/navigation';

const features = [
  {
    Icon: Workflow,
    title: "Workflow Automation",
    description: "Streamline your deployment process with automated workflows and CI/CD pipelines"
  },
  {
    Icon: Shield,
    title: "Enterprise Security",
    description: "Built-in security features with HashiCorp Vault integration and role-based access control"
  },
  {
    Icon: Database,
    title: "Configuration Management",
    description: "Centralized configuration management with GitOps practices"
  }
];

const technologies = [
  { name: 'ArgoCD', logo: '/logos/ArgoCD.svg' },
  { name: 'Kubernetes', logo: '/logos/Kubernetes.svg' },
  { name: 'ExternalSecret', logo: '/logos/eso-logo-large.png' },
  { name: 'Vault', logo: '/logos/HashiCorpVault.svg' },
  { name: 'Argo Workflow', logo: '/logos/argo-horizontal-color.png' },
];

const sections = [
  {
    title: "Secure by Default",
    description: "Built with security in mind, integrating HashiCorp Vault for secrets management and providing fine-grained access control.",
    image: "/logos/security-gopher.svg",
    features: [
      "Vault Integration",
      "RBAC Support",
      "Audit Logging",
      "Secret Rotation"
    ]
  },
  {
    title: "GitOps Ready",
    description: "Embrace GitOps practices with our integrated ArgoCD support, making deployment and configuration management a breeze.",
    image: "/logos/gitops-gopher.svg",
    features: [
      "ArgoCD Integration",
      "Git-based Workflows",
      "Automated Sync",
      "Drift Detection"
    ]
  },
  {
    title: "Enterprise Scale",
    description: "Built to scale with your organization, supporting multiple teams, projects, and environments.",
    image: "/logos/scale-gopher.svg",
    features: [
      "Multi-tenant Support",
      "Resource Quotas",
      "Team Management",
      "Environment Isolation"
    ]
  }
];

const faqs = [
  {
    question: "What is SquidFlow Platform?",
    answer: "SquidFlow Platform is a modern cloud-native application platform that combines ArgoCD and External Secrets to provide seamless deployment, security, and configuration management capabilities for your applications."
  },
  {
    question: "How does GitOps work with SquidFlow Platform?",
    answer: "SquidFlow Platform follows GitOps principles by using Git as the single source of truth. Changes to your application configurations are made through Git, and ArgoCD automatically syncs these changes to your Kubernetes clusters."
  },
  {
    question: "How does SquidFlow Platform handle secrets?",
    answer: "SquidFlow Platform integrates with HashiCorp Vault through External Secrets Operator to securely manage and inject secrets into your applications. This ensures that sensitive information is never stored in Git repositories."
  },
  {
    question: "What environments are supported?",
    answer: "SquidFlow Platform supports multiple environments (SIT, UAT, PRD) with environment-specific configurations and security policies. You can easily manage deployments across all environments from a single interface."
  },
  {
    question: "Can I integrate SquidFlow Platform with my existing CI/CD pipeline?",
    answer: "Yes, SquidFlow Platform is designed to work with your existing CI/CD tools. It can be integrated with popular CI platforms like Jenkins, GitHub Actions, and GitLab CI."
  },
  {
    question: "How does SquidFlow Platform ensure security?",
    answer: "SquidFlow Platform provides multiple security layers including RBAC, audit logging, secret rotation, and integration with enterprise security tools. All operations are logged and traceable."
  }
];

export default function Home() {
  const router = useRouter();

  return (
    <div className="flex flex-col min-h-screen">
      <Header isLoggedIn={false} />

      <section className="relative min-h-[90vh] flex items-center overflow-hidden bg-gradient-to-b from-background via-background/95 to-background/90">
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute -top-1/2 -right-1/2 w-[100rem] h-[100rem] bg-gradient-radial from-primary/5 via-transparent to-transparent opacity-50 blur-3xl" />
          <div className="absolute -bottom-1/2 -left-1/2 w-[100rem] h-[100rem] bg-gradient-radial from-blue-500/5 via-transparent to-transparent opacity-50 blur-3xl" />
        </div>

        <div className="container mx-auto px-4 relative z-10">
          <div className="flex flex-col md:flex-row items-center justify-between gap-16">
            <div className="w-full md:w-1/2 space-y-8">
              <div className="space-y-4">
                <h1 className="text-5xl md:text-7xl font-bold leading-tight">
                  <span className="bg-clip-text text-transparent bg-gradient-to-r from-primary via-blue-600 to-purple-600">
                    SquidFlow Platform
                  </span>
                  <br />
                  <span className="text-foreground">
                    for Cloud Native Apps
                  </span>
                </h1>
                <p className="text-xl md:text-2xl text-muted-foreground leading-relaxed max-w-2xl">
                  SquidFlow Platform combines ArgoCD and External Secrets to provide a seamless, secure, and scalable deployment experience.
                </p>
              </div>
              <div className="flex flex-col sm:flex-row gap-4">
                <Button
                  size="lg"
                  className="bg-primary/90 hover:bg-primary shadow-lg hover:shadow-xl transition-all duration-300 text-lg px-8"
                  onClick={() => router.push('/login')}
                >
                  Get Started
                </Button>
                <Button
                  size="lg"
                  variant="outline"
                  className="shadow-md hover:shadow-lg transition-all duration-300 text-lg px-8"
                >
                  Documentation
                </Button>
              </div>
            </div>

            <div className="w-full md:w-1/2 aspect-square">
              <Dog3D />
            </div>
          </div>
        </div>
      </section>

      <section className="py-32 bg-gradient-to-b from-background/95 to-background relative overflow-hidden">
        <div className="container mx-auto px-4">
          <div className="text-center space-y-4 mb-20">
            <h2 className="text-4xl md:text-5xl font-bold">
              Everything you need for modern deployments
            </h2>
            <p className="text-xl text-muted-foreground max-w-3xl mx-auto">
              Built for modern cloud-native applications with enterprise-grade security and scalability in mind.
            </p>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {features.map((feature, index) => (
              <Card key={index} className="bg-background/50 backdrop-blur-sm border-primary/10 hover:border-primary/20 transition-all duration-300 group">
                <CardHeader>
                  <div className="mb-6 p-3 rounded-lg bg-primary/5 w-fit group-hover:bg-primary/10 transition-colors">
                    <feature.Icon className="h-6 w-6 text-primary" />
                  </div>
                  <CardTitle className="text-xl">{feature.title}</CardTitle>
                </CardHeader>
                <CardContent className="text-muted-foreground">
                  {feature.description}
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {sections.map((section, index) => (
        <section key={index} className={`py-32 ${index % 2 === 0 ? 'bg-background' : 'bg-background/80'} relative overflow-hidden`}>
          <div className="container mx-auto px-4">
            <div className={`flex flex-col ${index % 2 === 0 ? 'md:flex-row' : 'md:flex-row-reverse'} items-center gap-16`}>
              <div className="w-full md:w-1/2 space-y-8">
                <div className="space-y-4">
                  <h2 className="text-4xl font-bold">{section.title}</h2>
                  <p className="text-xl text-muted-foreground">{section.description}</p>
                </div>
                <ul className="space-y-4">
                  {section.features.map((feature, i) => (
                    <li key={i} className="flex items-center gap-3 text-lg">
                      <div className="p-1.5 rounded-full bg-primary/10">
                        <Shield className="h-4 w-4 text-primary" />
                      </div>
                      <span>{feature}</span>
                    </li>
                  ))}
                </ul>
              </div>
              <div className="w-full md:w-1/2">
                <div className="relative aspect-square group">
                  <div className="absolute inset-0 bg-gradient-radial from-primary/5 via-transparent to-transparent opacity-50 group-hover:opacity-75 transition-opacity duration-500" />
                  <Image
                    src={section.image}
                    alt={section.title}
                    width={500}
                    height={500}
                    className="object-contain relative z-10 transition-transform duration-500 group-hover:scale-105"
                    priority={index === 0}
                  />
                </div>
              </div>
            </div>
          </div>
        </section>
      ))}

      <section className="py-24 bg-background/90 backdrop-blur-sm relative overflow-hidden">
        <div className="container mx-auto px-4">
          <h2 className="text-3xl font-bold text-center mb-16">
            Powered by Industry Leading Technologies
          </h2>
          <div className="w-full overflow-hidden">
            <div className="flex animate-scroll">
              {[...technologies, ...technologies].map((tech, index) => (
                <div key={index} className="flex-shrink-0 w-32 mx-8 hover:scale-110 transition-transform duration-300">
                  <Image
                    src={tech.logo}
                    alt={tech.name}
                    width={100}
                    height={100}
                    className="object-contain filter grayscale hover:grayscale-0 transition-all duration-300"
                    priority={index < technologies.length}
                  />
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* FAQ Section */}
      <section className="py-24 bg-background/90 backdrop-blur-sm relative overflow-hidden">
        <div className="container mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold mb-4">Frequently Asked Questions</h2>
            <p className="text-xl text-muted-foreground">
              Everything you need to know about SquidFlow Platform
            </p>
          </div>

          <div className="max-w-3xl mx-auto">
            <Accordion type="single" collapsible className="w-full space-y-4">
              {faqs.map((faq, index) => (
                <AccordionItem
                  key={index}
                  value={`item-${index}`}
                  className="bg-card border rounded-lg px-6 shadow-sm hover:shadow-md transition-all duration-200"
                >
                  <AccordionTrigger className="text-left hover:no-underline py-6">
                    <div className="flex items-center space-x-3">
                      <span className="flex-shrink-0 w-8 h-8 flex items-center justify-center rounded-full bg-primary/10 text-primary">
                        {index + 1}
                      </span>
                      <span className="text-lg font-medium">{faq.question}</span>
                    </div>
                  </AccordionTrigger>
                  <AccordionContent className="text-muted-foreground pb-6 pt-2 pl-11">
                    {faq.answer}
                  </AccordionContent>
                </AccordionItem>
              ))}
            </Accordion>

            {/* Call to Action */}
            <div className="text-center mt-16">
              <p className="text-lg text-muted-foreground mb-8">
                Still have questions? We&apos;re here to help.
              </p>
              <div className="flex justify-center gap-4">
                <Button
                  variant="outline"
                  className="shadow-md hover:shadow-lg transition-all duration-300"
                >
                  Contact Support
                </Button>
                <Button
                  className="bg-primary/90 hover:bg-primary shadow-lg hover:shadow-xl transition-all duration-300"
                >
                  Read Documentation
                </Button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <Footer />
    </div>
  );
}