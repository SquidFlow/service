import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FileText } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { ApplicationTemplate } from '@/types/application';

interface GeneralInfoProps {
  app: ApplicationTemplate;
}

export function GeneralInfo({ app }: GeneralInfoProps) {
  return (
    <Card className="col-span-2 bg-white dark:bg-gray-800 shadow-sm hover:shadow-md transition-shadow duration-200">
      <CardHeader>
        <CardTitle className="flex items-center space-x-2">
          <FileText className="h-5 w-5 text-blue-500" />
          <span>General Information</span>
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="grid grid-cols-2 gap-6">
          <div className="space-y-4">
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Owner
              </p>
              <div className="flex items-center space-x-2 mt-1">
                <Avatar className="h-6 w-6">
                  <AvatarFallback className="bg-blue-100 text-blue-600">
                    {app.owner?.split(" ").map((n: string) => n[0]).join("") || app.owner?.[0] || 'U'}
                  </AvatarFallback>
                </Avatar>
                <span className="font-medium">{app.owner}</span>
              </div>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400">
                Last Update
              </p>
              <p className="mt-1">
                {new Date(app.lastUpdate).toLocaleString()}
              </p>
            </div>
          </div>
        </div>

        <div className="pt-4 border-t border-gray-200 dark:border-gray-700">
          <h4 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
            Remote Repository
          </h4>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-gray-500">Repository URL</span>
              <a
                href={app.template?.source?.url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm text-blue-500 hover:text-blue-600 font-mono"
              >
                {app.template?.source?.url}
              </a>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}