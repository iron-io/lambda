package example;

import com.amazonaws.services.lambda.runtime.Context;

public class Hello {
    public String myHandler(int myCount, Context context) {
        System.out.println("Hello \n world");
        System.out.println("Context: " + context.getClientContext());
        System.out.println("Function name: " + context.getFunctionName());
        System.out.println("Time remaining in milliseconds: " + context.getRemainingTimeInMillis());
        return String.valueOf(myCount);
    }
    public String test(int myVar) {
        System.out.println("Hello \n world");
        System.out.println(String.valueOf(myVar));
        return String.valueOf(myVar);
    }
}
