package example;

import com.amazonaws.services.lambda.runtime.Context;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;

public class Hello {
    public String myHandlerInt(int myCount, Context context) {
        System.out.println("Hello \n world");
        System.out.println("Context: " + context.getClientContext());
        System.out.println("Function name: " + context.getFunctionName());
        System.out.println("Time remaining in milliseconds: " + context.getRemainingTimeInMillis());
        return String.valueOf(myCount);
    }

    public String myHandlerString(String text, Context context) {
        System.out.println("Hello \n world");
        System.out.println("Context: " + context.getClientContext());
        System.out.println("Function name: " + context.getFunctionName());
        System.out.println("Time remaining in milliseconds: " + context.getRemainingTimeInMillis());
        return text;
    }

    public static void myHandlerIO(InputStream inputStream, OutputStream outputStream, Context context) throws IOException {
        int letter;
        while((letter = inputStream.read()) != -1)
        {
            outputStream.write(Character.toUpperCase(letter));
        }
    }

    // Define two classes/POJOs for use with Lambda function.
    public static class RequestClass {
        public String firstName;
        public String lastName;

        public String getFirstName() {
            return firstName;
        }

        public void setFirstName(String firstName) {
            this.firstName = firstName;
        }

        public String getLastName() {
            return lastName;
        }

        public void setLastName(String lastName) {
            this.lastName = lastName;
        }

        public RequestClass(String firstName, String lastName) {
            this.firstName = firstName;
            this.lastName = lastName;
        }

        public RequestClass() {
        }
    }

    public static class ResponseClass {
        String greetings;
        String greetings2 = "123";
        String greetings3 = "123";

        public String getGreetings() {
            return greetings;
        }
        public String getGreetings3() {
            return greetings3;
        }

        public void setGreetings(String greetings) {
            this.greetings = greetings;
        }
        public void setGreetings3(String greetings) {
            this.greetings3 = greetings;
        }

        public ResponseClass(String greetings) {
            this.greetings = greetings;
        }

        public ResponseClass() {
        }

    }

    public static ResponseClass myHandlerPOJO(RequestClass request, Context context){
        String greetingString = "Hello %s, %s." + request.firstName + request.lastName;
        return new ResponseClass(greetingString);
    }

}