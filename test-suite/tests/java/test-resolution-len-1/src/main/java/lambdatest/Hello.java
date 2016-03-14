package lambdatest;

import com.amazonaws.services.lambda.runtime.Context;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class Hello {
    // This won't get called because only stream handlers can have 2-3 args.
    public static void myHandler(int payload, int another, Context context) {
        int a = 1/0;
        System.out.println("This does not get called.");
    }

    public static void myHandler(int payload, Context context) {
        System.out.println("This gets called.");
    }

    public static void myHandler(Context context) {
        int a = 1/0;
        System.out.println("This does not get called.");
    }

    public static void myHandler() {
        int a = 1/0;
        System.out.println("This does not get called.");
    }
}
