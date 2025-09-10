import { useState } from 'react'
import { CreateDagFormSchema, type CreateDagForm } from '../types'

export default function Home() {
  const [form, setForm] = useState<CreateDagForm>({
    name: '',
    description: '',
  })
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isValid, setIsValid] = useState(false)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    
    const result = CreateDagFormSchema.safeParse(form)
    
    if (result.success) {
      setErrors({})
      setIsValid(true)
      console.log('Form is valid:', result.data)
      // Here you would typically send the data to your backend
    } else {
      const formErrors: Record<string, string> = {}
      result.error.errors.forEach((err) => {
        if (err.path[0]) {
          formErrors[err.path[0] as string] = err.message
        }
      })
      setErrors(formErrors)
      setIsValid(false)
    }
  }

  const handleInputChange = (field: keyof CreateDagForm, value: string) => {
    setForm(prev => ({ ...prev, [field]: value }))
    setIsValid(false)
    
    // Clear error for this field when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }))
    }
  }

  return (
    <div className="max-w-md mx-auto">
      <h1 className="text-2xl font-bold text-gray-900 mb-6">
        Create New DAG
      </h1>
      
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="name" className="block text-sm font-medium text-gray-700">
            Name
          </label>
          <input
            type="text"
            id="name"
            value={form.name}
            onChange={(e) => handleInputChange('name', e.target.value)}
            className={`mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 ${
              errors.name ? 'border-red-500' : ''
            }`}
            placeholder="Enter DAG name"
          />
          {errors.name && (
            <p className="mt-1 text-sm text-red-600">{errors.name}</p>
          )}
        </div>

        <div>
          <label htmlFor="description" className="block text-sm font-medium text-gray-700">
            Description (optional)
          </label>
          <textarea
            id="description"
            value={form.description || ''}
            onChange={(e) => handleInputChange('description', e.target.value)}
            rows={3}
            className={`mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 ${
              errors.description ? 'border-red-500' : ''
            }`}
            placeholder="Enter DAG description"
          />
          {errors.description && (
            <p className="mt-1 text-sm text-red-600">{errors.description}</p>
          )}
        </div>

        <button
          type="submit"
          className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
        >
          Create DAG
        </button>
      </form>

      {isValid && (
        <div className="mt-4 p-4 bg-green-100 border border-green-400 text-green-700 rounded">
          ✅ Form validation passed! Check console for the validated data.
        </div>
      )}
      
      <div className="mt-8 p-4 bg-blue-50 border border-blue-200 rounded">
        <h3 className="text-sm font-medium text-blue-900 mb-2">
          This form demonstrates:
        </h3>
        <ul className="text-sm text-blue-800 list-disc list-inside space-y-1">
          <li>Zod schema validation</li>
          <li>TypeScript type safety</li>
          <li>Real-time error handling</li>
          <li>Tailwind CSS styling</li>
        </ul>
      </div>
    </div>
  )
}
