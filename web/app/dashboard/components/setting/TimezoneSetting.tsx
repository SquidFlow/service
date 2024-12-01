import { Card, CardHeader, CardContent, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

interface TimezoneSettingProps {
  defaultTimezone?: string;
  onTimezoneChange?: (timezone: string) => void;
}

export function TimezoneSetting({ defaultTimezone = "Asia/Shanghai", onTimezoneChange }: TimezoneSettingProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>System Settings</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <Label>Timezone</Label>
            <Select
              defaultValue={defaultTimezone}
              onValueChange={onTimezoneChange}
            >
              <SelectTrigger className="w-[280px]">
                <SelectValue placeholder="Select timezone" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="Asia/Shanghai">China Standard Time (UTC+8)</SelectItem>
                <SelectItem value="Asia/Singapore">Singapore Time (UTC+8)</SelectItem>
                <SelectItem value="Asia/Kolkata">India Standard Time (UTC+5:30)</SelectItem>
                <SelectItem value="UTC">Coordinated Universal Time (UTC+0)</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}