"use client"

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, LineChart, Line, XAxis, YAxis, CartesianGrid, Legend } from 'recharts'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { resourceQuota, usageTrends, aggregatedResources, nodes } from './resourcePoolMock';
import { ResourceQuotaInterface } from '../interfaces';

export function ResourcePool({ activeSubMenu }: { activeSubMenu: string }) {
  const [aggregationType, setAggregationType] = useState<'tenant' | 'appcode' | 'application'>('tenant');

  const renderResourceOverview = () => {
    const parseValue = (value: string) => {
      const numericValue = parseFloat(value);
      if (isNaN(numericValue)) return 0;
      if (value.includes('Gi')) return numericValue * 1024;
      if (value.includes('Mi')) return numericValue;
      if (value.includes('cores')) return numericValue * 1000;
      return numericValue;
    };

    const data = [
      { name: 'CPU', value: parseValue(resourceQuota.cpu), color: '#0088FE' },
      { name: 'Memory', value: parseValue(resourceQuota.memory), color: '#00C49F' },
      { name: 'Storage', value: parseValue(resourceQuota.storage), color: '#FFBB28' }
    ];

    return (
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {data.map((item) => (
            <Card key={item.name}>
              <CardHeader>
                <CardTitle>{item.name} Usage</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="h-[200px]">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={[item]}
                        cx="50%"
                        cy="50%"
                        innerRadius={60}
                        outerRadius={80}
                        fill={item.color}
                        paddingAngle={5}
                        dataKey="value"
                      >
                        <Cell key={`cell-${item.name}`} fill={item.color} />
                      </Pie>
                      <Tooltip formatter={(value) => [`${value} ${item.name === 'CPU' ? 'm' : 'Mi'}`, item.name]} />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
                <div className="mt-4 text-center">
                  <p className="font-semibold">
                    {resourceQuota[item.name.toLowerCase() as 'cpu' | 'memory' | 'storage']}
                  </p>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
        <Card>
          <CardHeader>
            <CardTitle>Resource Usage Trends</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-[400px]">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={usageTrends}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Line type="monotone" dataKey="cpu" stroke="#0088FE" name="CPU (cores)" />
                  <Line type="monotone" dataKey="memory" stroke="#00C49F" name="Memory (GB)" />
                  <Line type="monotone" dataKey="storage" stroke="#FFBB28" name="Storage (GB)" />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <div className="flex justify-between items-center">
              <CardTitle>Aggregated Resource Usage</CardTitle>
              <Select value={aggregationType} onValueChange={(value: 'tenant' | 'appcode' | 'application') => setAggregationType(value)}>
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="Select aggregation type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="tenant">By Tenant</SelectItem>
                  <SelectItem value="appcode">By App Code</SelectItem>
                  <SelectItem value="application">By Application</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>CPU</TableHead>
                  <TableHead>Memory</TableHead>
                  <TableHead>Storage</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {aggregationType === 'tenant' && aggregatedResources.tenants.map((item, index) => (
                  <TableRow key={index}>
                    <TableCell>{item.name}</TableCell>
                    <TableCell>{item.cpu}</TableCell>
                    <TableCell>{item.memory}</TableCell>
                    <TableCell>{item.storage}</TableCell>
                  </TableRow>
                ))}
                {aggregationType === 'appcode' && aggregatedResources.appcodes.map((item, index) => (
                  <TableRow key={index}>
                    <TableCell>{item.name}</TableCell>
                    <TableCell>{item.cpu}</TableCell>
                    <TableCell>{item.memory}</TableCell>
                    <TableCell>{item.storage}</TableCell>
                  </TableRow>
                ))}
                {aggregationType === 'application' && aggregatedResources.applications.map((item, index) => (
                  <TableRow key={index}>
                    <TableCell>{item.name}</TableCell>
                    <TableCell>{item.cpu}</TableCell>
                    <TableCell>{item.memory}</TableCell>
                    <TableCell>{item.storage}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    );
  };

  const renderNodes = () => {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Cluster Nodes</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>IP</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>CPU</TableHead>
                <TableHead>Memory</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {nodes.map((node) => (
                <TableRow key={node.id}>
                  <TableCell>{node.name}</TableCell>
                  <TableCell>{node.ip}</TableCell>
                  <TableCell>
                    <span className={`px-2 py-1 rounded-full text-sm ${
                      node.status === 'Ready'
                        ? 'bg-green-100 text-green-800'
                        : 'bg-red-100 text-red-800'
                    }`}>
                      {node.status}
                    </span>
                  </TableCell>
                  <TableCell>{node.cpu}</TableCell>
                  <TableCell>{node.memory}</TableCell>
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
      {activeSubMenu === 'Overview' && renderResourceOverview()}
      {activeSubMenu === 'Nodes' && renderNodes()}
    </>
  );
}

