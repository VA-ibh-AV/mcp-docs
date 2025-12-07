"use client";
import { FcGoogle } from "react-icons/fc";
import { Label } from "../ui/label";
import { Button } from "../ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Card,
  CardHeader,
  CardTitle,
  CardContent,
  CardFooter,
  CardDescription,
} from "../ui/card";
import { Input } from "../ui/input";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { SignInSchema, signInSchema } from "@/schemas/signIn.schema";
import { Eye, EyeOff } from "lucide-react";
import { useState } from "react";
import { useRouter } from "next/navigation";

const LoginForm = () => {
  const [showPassword, setShowPassword] = useState(false);
  const router = useRouter();

  const form = useForm<SignInSchema>({
    resolver: zodResolver(signInSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  const onSubmit = async (data: SignInSchema) => {
    console.log(data);
    
    // Set the access token in cookies for testing purposes
    // Note: httpOnly cookies cannot be set from client-side, so using regular cookie for testing
    const maxAge = 60 * 60 * 24 * 30; // 30 days
    const expires = new Date(Date.now() + maxAge * 1000).toUTCString();
    const secure = process.env.NODE_ENV === "production" ? "Secure;" : "";
    document.cookie = `access_token=test_access_token; expires=${expires}; path=/; ${secure}SameSite=Lax`;
    
    // Redirect user to /dashboard
    router.push("/dashboard");
  };

  return (
    <Card className="w-full max-w-sm border-0 shadow-lg">
      <CardHeader>
        <CardTitle>Welcome back!</CardTitle>
        <CardDescription>
          Enter your email and password to login to your account.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Email</FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Password</FormLabel>

                  <FormControl>
                    <div className="relative">
                      <Input
                        {...field}
                        type={showPassword ? "text" : "password"}
                        className="pr-10"
                      />

                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="absolute right-2 top-1/2 -translate-y-1/2 h-7 w-7"
                        onClick={() => setShowPassword((v) => !v)}
                      >
                        {showPassword ? (
                          <EyeOff className="h-4 w-4" />
                        ) : (
                          <Eye className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </FormControl>

                  <FormMessage />
                </FormItem>
              )}
            />

            <Button type="submit" className="w-full">
              Login
            </Button>
          </form>
        </Form>
      </CardContent>
      <CardFooter className="flex-col gap-2">
        <div className="flex items-center w-full my-4">
          <div className="flex-1 border-t border-slate-400" />
          <span className="px-3 text-sm text-slate-500 bg-card">
            Or continue with
          </span>
          <div className="flex-1 border-t border-slate-400" />
        </div>

        <Button variant="outline" className="w-full flex gap">
          <FcGoogle size={18} />
          Login with Google
        </Button>
      </CardFooter>
    </Card>
  );
};

export default LoginForm;
