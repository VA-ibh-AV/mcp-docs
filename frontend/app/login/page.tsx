import LoginForm from "@/components/authentication/LoginForm";
import RegisterForm from "@/components/authentication/RegisterForm";
import { Logo } from "@/components/ui/logo";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"


const LoginPage = () => {
  return (
    <div className="flex w-full max-w-sm mx-auto flex-col gap-6 justify-center items-center h-screen">

    <Logo />
        
    <Tabs defaultValue="login" className="w-full">

    <TabsList className="bg-card p-1 w-full">
          <TabsTrigger 
            value="login" 
            className="flex-1 data-[state=active]:bg-primary data-[state=active]:text-primary-foreground data-[state=active]:shadow-sm transition-all"
          >
            Login
          </TabsTrigger>
          <TabsTrigger 
            value="register" 
            className="flex-1 data-[state=active]:bg-primary data-[state=active]:text-primary-foreground data-[state=active]:shadow-sm transition-all"
          >
            Register
          </TabsTrigger>
        </TabsList>
        <TabsContent value="login">
          <LoginForm />
        </TabsContent>
        <TabsContent value="register">
          <RegisterForm />
        </TabsContent>
    </Tabs>
    </div>
  );
};

export default LoginPage;
