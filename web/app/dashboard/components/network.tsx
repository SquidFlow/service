"use client"

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { VirtualService } from '../interfaces';

// Mock data for VirtualServices
const virtualServices: VirtualService[] = [
  { id: 1, name: 'frontend-vs', hosts: ['frontend.example.com'], gateways: ['main-gateway'], createdAt: '2023-05-01' },
  { id: 2, name: 'backend-vs', hosts: ['api.example.com'], gateways: ['main-gateway'], createdAt: '2023-05-02' },
  { id: 3, name: 'monitoring-vs', hosts: ['grafana.example.com', 'prometheus.example.com'], gateways: ['monitoring-gateway'], createdAt: '2023-05-03' },
];

export function NetworkMenu({ activeSubMenu }: { activeSubMenu: string }) {
  const renderVirtualServices = () => {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Virtual Services</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Hosts</TableHead>
                <TableHead>Gateways</TableHead>
                <TableHead>Created At</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {virtualServices.map((vs) => (
                <TableRow key={vs.id}>
                  <TableCell>{vs.name}</TableCell>
                  <TableCell>{vs.hosts.join(', ')}</TableCell>
                  <TableCell>{vs.gateways.join(', ')}</TableCell>
                  <TableCell>{vs.createdAt}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    );
  };

  return (
    <>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">{activeSubMenu}</h1>
      </div>
      {activeSubMenu === 'VirtualService' && renderVirtualServices()}
    </>
  );
}