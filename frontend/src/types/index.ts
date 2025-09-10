import { z } from 'zod'

// Example schema for a DAG (Directed Acyclic Graph) to match the backend
export const DagSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1, 'Name is required'),
  description: z.string().optional(),
  nodes: z.array(z.object({
    id: z.string(),
    name: z.string(),
    type: z.string(),
  })),
  edges: z.array(z.object({
    from: z.string(),
    to: z.string(),
  })),
  createdAt: z.string().datetime().optional(),
  updatedAt: z.string().datetime().optional(),
})

export type Dag = z.infer<typeof DagSchema>

// User schema example
export const UserSchema = z.object({
  id: z.string().uuid(),
  name: z.string().min(1, 'Name is required'),
  email: z.string().email('Invalid email'),
})

export type User = z.infer<typeof UserSchema>

// Form validation schemas
export const CreateDagFormSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name too long'),
  description: z.string().max(500, 'Description too long').optional(),
})

export type CreateDagForm = z.infer<typeof CreateDagFormSchema>
